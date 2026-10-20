# Despliegue en Vercel

## 1. Diagnóstico: por qué no era desplegable tal cual

El repositorio original era una app Gin válida en local, pero incompatible con Vercel:

| Problema | Consecuencia en Vercel |
| --- | --- |
| `main.go` fijaba `:8081` | Vercel inyecta `PORT`; el proceso no escuchaba donde Vercel enruta. |
| `repository.NewInMemoryRepository()` en el router | Usuarios y sesiones se perdían en cada cold start y no se compartían entre instancias. |
| `r.LoadHTMLGlob("templates/*")` y `r.Static("./static")` | Dependencia del filesystem de ejecución, efímero y no escribible. |
| Administrador `admin@automarket.com` / `admin123` hardcodeado | Secreto en el repositorio y riesgo crítico en producción. |
| Sin configuración por entorno | No había forma segura de inyectar base de datos, cookies ni secretos. |
| Router construía repositorios y servicios | `main` y HTTP conocían implementaciones concretas; difícil de testear y sustituir. |
| `ShowProfile` devolvía `domain.User` completo | Filtraba `Password`/`Salt` por JSON. |
| Hash de contraseñas iterado a mano con 20.000 iteraciones | Coste insuficiente y formato no versionado. |
| Logs a archivo (ADR-0005) | El filesystem de la función es efímero; los logs deben ir a stdout. |
| `/dashboard` y formularios sin cookies seguras | La cookie no tenía `Secure`/`SameSite` explícitos. |

Vercel **sí soporta Go** actualmente mediante el Framework Preset `go`: detecta `go.mod` y un entrypoint `main.go`, ejecuta el binario y le pasa `PORT`. También soporta Gin. La app podía adaptarse sin reescribir la lógica en otro lenguaje.

## 2. Solución elegida

**Mantener el servidor Go/Gin y desplegarlo con el preset `go` de Vercel, usando PostgreSQL gestionado para persistencia.**

Cambios estructurales:

1. `main.go` carga configuración validada, construye la app y escucha en `:$PORT` con apagado graceful.
2. `internal/app` es el composition root; `router` ya no instancia repositorios.
3. PostgreSQL sustituye los mapas en memoria. Usuarios y sesiones persisten entre invocaciones.
4. Templates y CSS se embeben con `go:embed`; el runtime no lee archivos locales.
5. Las sesiones se guardan como fingerprint SHA-256, con expiración y cookie `HttpOnly`, `Secure` y `SameSite=Lax`.
6. Las contraseñas usan PBKDF2-HMAC-SHA256 con 600.000 iteraciones y formato versionado.
7. Los secretos y la URL de base de datos se leen de variables de entorno; no hay credenciales en el código.
8. Las respuestas usan view models sin `PasswordHash`.

### Alternativas descartadas

| Alternativa | Motivo |
| --- | --- |
| Funciones Go en `api/*.go` | Obliga a un adaptador por entrypoint, complica el routing de todas las páginas y no aporta ventajas frente al preset de servidor. |
| Frontend estático + API Go en otro proveedor | Válida, pero añade dos despliegues, CORS y configuración de dominio sin necesidad. |
| SQLite o JSON en el proyecto | El filesystem de Vercel es efímero y no apto para escritura persistente. |
| Mantener memoria | Pierde datos y sesiones; inviable en producción. |
| Servicio externo (Render/Fly/Railway) | Buena opción si se necesita un proceso siempre activo, pero Vercel ya ejecuta el servidor Go. |

## 3. Configuración de Vercel

`vercel.json`:

```json
{
  "$schema": "https://openapi.vercel.sh/vercel.json",
  "framework": "go"
}
```

Notas:

- `go.mod` debe estar en la raíz. Lo está.
- El entrypoint detectado es `main.go`.
- No se fija el puerto: el código usa `PORT`.
- No se necesita `buildCommand`: Vercel ejecuta el build de Go.
- Las dependencias se resuelven con `go.mod`/`go.sum`.
- La región por defecto es `iad1`. Si la base de datos está en otra región, cámbiala en **Project Settings → Functions → Region** para reducir latencia. La región debe coincidir con la de PostgreSQL.
- El runtime Go de Vercel está en **Beta**. Fija la versión con la directiva `go` de `go.mod` (actualmente `1.25.0`).

## 4. Base de datos

Opciones válidas: Neon, Supabase, Vercel Marketplace (Postgres) o cualquier PostgreSQL accesible por TLS.

Recomendaciones:

- Usa la **URL pooled** del proveedor para el runtime (Neon pooler, Supabase pooler o PgBouncer). Vercel puede abrir varias instancias y cada una mantiene su propio pool.
- Usa `sslmode=require` (o `verify-full` si el proveedor lo soporta).
- Deja `DB_MAX_OPEN_CONNS=5` y `DB_MAX_IDLE_CONNS=2` como punto de partida. Súbelos solo midiendo.
- Co-loca la región de la función y la de la base de datos.

El esquema se crea automáticamente al arrancar:

```
users(id, name, email unique, password_hash, role, created_at)
sessions(token_hash primary key, user_id fk, created_at, expires_at)
```

La migración es idempotente y usa `pg_advisory_xact_lock`, por lo que varios cold starts concurrentes no chocan.

## 5. Variables de entorno en Vercel

Configúralas en **Project Settings → Environment Variables** (Production y Preview según corresponda).

| Variable | Producción | Notas |
| --- | --- | --- |
| `DATABASE_URL` | Obligatoria | URL pooled con TLS. Marcar como *Sensitive*. |
| `APP_ENV` | `production` | Habilita la configuración de producción. Si se omite, `VERCEL=1` la activa automáticamente. |
| `COOKIE_SECURE` | `true` | Vercel sirve HTTPS. |
| `SESSION_TTL` | `24h` | Ajusta según el negocio. |
| `ADMIN_NAME` | Opcional | Default `Administrador`. |
| `ADMIN_EMAIL` | Opcional | Requiere `ADMIN_PASSWORD`. |
| `ADMIN_PASSWORD` | Opcional | Mínimo 12 caracteres. Marcar como *Sensitive*. |
| `LOG_LEVEL` | `info` | `debug` solo temporalmente. |
| `DB_MAX_OPEN_CONNS` | `5` | Ajustar al límite del proveedor. |
| `DB_MAX_IDLE_CONNS` | `2` | — |
| `DB_CONN_MAX_LIFETIME` | `5m` | — |
| `DB_CONNECT_TIMEOUT` | `10s` | — |
| `PORT` | **No definir** | Vercel lo inyecta. |
| `COOKIE_NAME` | Opcional | Default `session_token`. |

Si usas la integración de Vercel/Neon y esta define `POSTGRES_URL`, el código la acepta como fallback. Aun así, `DATABASE_URL` explícita es la opción recomendada.

### Bootstrap del administrador

Si `ADMIN_EMAIL` y `ADMIN_PASSWORD` están definidos, al arrancar se crea el administrador **solo si el email no existe**. Es idempotente y nunca promueve una cuenta existente: si el email ya pertenece a un vendedor, el arranque falla con un error explícito para evitar una configuración ambigua. Tras crear el administrador puedes retirar `ADMIN_PASSWORD` del entorno; la cuenta permanece en la base de datos. Para rotar la contraseña se necesita un flujo administrativo o una actualización manual controlada.

## 6. Despliegue

### Opción A: Git + dashboard

1. Sube los cambios a GitHub/GitLab/Bitbucket.
2. En Vercel, **Add New → Project** e importa el repositorio.
3. Verifica que el Framework Preset sea **Go** (lo fija `vercel.json`).
4. Añade las variables de entorno.
5. Deploy.

### Opción B: Vercel CLI

```bash
npm i -g vercel
vercel link
vercel env add DATABASE_URL production
vercel env add APP_ENV production
vercel env add COOKIE_SECURE production
vercel env add ADMIN_EMAIL production
vercel env add ADMIN_PASSWORD production
vercel --prod
```

No ejecutes `vercel dev` esperando que use el preset Go si tu CLI no lo soporta; para desarrollo local usa `go run .`.

## 7. Validación posterior al despliegue

1. **Salud y base de datos**

```bash
curl -i https://tu-proyecto.vercel.app/healthz
```

Esperado: `200` con `{"status":"ok"}`. Si devuelve `503`, la base de datos no es accesible.

2. **Landing y assets**

```bash
curl -I https://tu-proyecto.vercel.app/
curl -I https://tu-proyecto.vercel.app/static/css/style.css
```

Esperado: `200` y `text/html` / `text/css`.

3. **Registro y login**

- Abre `/register`, crea un vendedor.
- Inicia sesión en `/login`.
- Verifica en DevTools que la cookie `session_token` sea `HttpOnly`, `Secure` y `SameSite=Lax`.
- Abre `/dashboard`; debe mostrar el usuario.

4. **Autorización**

- Con rol vendedor, `/admin-panel` debe devolver `403`.
- Sin cookie, `/dashboard` debe devolver `401`.
- `/profile` no debe contener `PasswordHash` ni el hash de la contraseña.

5. **Persistencia**

- Redeploy o espera un cold start.
- Inicia sesión de nuevo: el usuario debe seguir existiendo y la sesión anterior debe seguir siendo válida hasta expirar.

6. **Logs**

En **Vercel → Logs** deben aparecer líneas JSON `http_request` con método, ruta, estado y duración. No deben aparecer contraseñas, tokens ni cookies.

## 8. Riesgos y limitaciones

- **Go runtime Beta:** el preset puede cambiar. Mantén `go.mod` y `vercel.json` versionados y prueba cada actualización. Si el preset se retirara, la app puede moverse a Render/Fly/Railway o a un contenedor sin cambiar la lógica: solo escucha `$PORT` y usa PostgreSQL.
- **Cold starts + PBKDF2:** 600.000 iteraciones son costosas por diseño. El primer login tras un cold start puede tardar más y consume CPU facturable. No bajes las iteraciones sin evaluar el riesgo; si el coste es crítico, considera Argon2id en el futuro.
- **Conexiones a base de datos:** cada instancia mantiene su pool. Usa el endpoint pooled y límites conservadores. Un pico de concurrencia puede agotar las conexiones del proveedor.
- **Limpieza de sesiones:** se ejecuta de forma oportunista en cada login. Para volúmenes altos, añade un cron de Vercel o un job que borre `expires_at <= now()`.
- **CSRF:** `SameSite=Lax` mitiga la mayoría de ataques, pero no sustituye un token CSRF en formularios autenticados. Añádelo antes de incorporar operaciones sensibles.
- **Sin rate limiting:** login y registro no tienen límite de intentos. Añade rate limiting en Vercel Firewall o en la capa HTTP.
- **Sin verificación de email, recuperación de contraseña ni MFA.**
- **`/users` y `POST /cars` siguen siendo placeholders.** No hay módulo de vehículos persistido.
- **Logs:** van a stdout; no hay retención propia ni rotación. Usa Vercel Logs/Drains si necesitas auditoría a largo plazo.
- **Filesystem:** cualquier archivo que se escriba en runtime se pierde. No añadas escritura local.
- **Límites de Vercel:** payload de 4,5 MB, duración máxima de 300 s por defecto en Hobby. No afectan al flujo actual, pero condicionan futuras subidas de imágenes.
- **Una sola región:** la app no está diseñada para escritura multi-región. PostgreSQL sigue siendo el punto central.

## 9. Rollback

Vercel permite volver a un deployment anterior desde el dashboard. El esquema es aditivo (`CREATE TABLE IF NOT EXISTS`) y no destructivo; un rollback de aplicación no elimina datos. Si en el futuro se añaden migraciones destructivas, versiona el esquema y aplica una estrategia expand/contract.
