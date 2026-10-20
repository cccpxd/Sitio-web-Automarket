# AutoMarket

Aplicación web de venta de autos usados. Incluye landing page, registro de vendedores, inicio de sesión con sesiones en servidor y rutas protegidas por rol (`seller` y `admin`).

## Stack

- Go 1.25+
- Gin
- PostgreSQL (driver `pgx` v5 sobre `database/sql`)
- Templates HTML y CSS embebidos con `go:embed`
- Vercel (Framework Preset `go`) para despliegue

## Arquitectura

```
main.go                          # entrypoint corto: carga config y arranca el servidor
internal/
  domain/                        # entidades, errores y reglas de validación puras
  service/                       # casos de uso + puertos (interfaces)
  repository/postgres/           # adaptadores de persistencia y migraciones
  security/                      # PBKDF2, tokens opacos e IDs
  httpapi/
    contracts.go                 # puertos consumidos por la capa HTTP
    handlers/                    # adaptadores de entrada (Gin)
    middleware/                  # autenticación, roles, logs y headers
    router/                      # definición de rutas
  views/                         # view models de presentación
  webui/                         # templates y estáticos embebidos
  app/                           # composition root
  config/                        # lectura y validación de variables de entorno
  platform/                      # servidor HTTP, logging y reloj
docs/
  DEPLOY_VERCEL.md               # guía de despliegue y validación
  ARCHITECTURE.md                # detalle de capas y dependencias
  ADRs/                          # decisiones de arquitectura
```

La regla de dependencias es unidireccional:

```
httpapi/adapters -> service -> domain
repository/security/webui -> implementan puertos de service o httpapi
app -> conecta todas las piezas
```

El dominio y los servicios no importan Gin, HTML, PostgreSQL ni leen variables de entorno.

## Ejecución local

Requisitos: Go 1.25+ y PostgreSQL.

1. Crea una base de datos local. Ejemplo con Docker:

```bash
docker run --name automarket-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=automarket \
  -p 5432:5432 \
  -d postgres:17
```

2. Exporta las variables (PowerShell):

```powershell
$env:DATABASE_URL="postgres://postgres:postgres@localhost:5432/automarket?sslmode=disable"
$env:COOKIE_SECURE="false"
$env:PORT="8080"
go run .
```

En Bash:

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/automarket?sslmode=disable"
export COOKIE_SECURE="false"
export PORT="8080"
go run .
```

3. Abre `http://localhost:8080`.

Las tablas se crean automáticamente al arrancar. Si defines `ADMIN_EMAIL` y `ADMIN_PASSWORD`, se crea un administrador inicial de forma idempotente.

## Variables de entorno

| Variable | Obligatoria | Default | Descripción |
| --- | --- | --- | --- |
| `DATABASE_URL` | Sí | — | URL de PostgreSQL. También acepta `POSTGRES_URL` de integraciones Vercel/Neon. |
| `APP_ENV` | No | `production` en Vercel, `development` local | `development` o `production`. En Vercel se detecta automáticamente. |
| `PORT` | No | `8080` | Vercel lo inyecta automáticamente. |
| `COOKIE_SECURE` | No | `true` | Usa `false` solo en HTTP local. |
| `SESSION_TTL` | No | `24h` | Duración de la sesión. |
| `COOKIE_NAME` | No | `session_token` | Nombre de la cookie de sesión. |
| `ADMIN_NAME` | No | `Administrador` | Nombre del administrador inicial. |
| `ADMIN_EMAIL` | No | — | Requiere `ADMIN_PASSWORD`. |
| `ADMIN_PASSWORD` | No | — | Mínimo 12 caracteres. Requiere `ADMIN_EMAIL`. |
| `LOG_LEVEL` | No | `info` | `debug`, `info`, `warn` o `error`. |
| `DB_MAX_OPEN_CONNS` | No | `5` | Conexiones abiertas máximas. |
| `DB_MAX_IDLE_CONNS` | No | `2` | Conexiones idle máximas. |
| `DB_CONN_MAX_LIFETIME` | No | `5m` | Vida máxima de una conexión. |
| `DB_CONNECT_TIMEOUT` | No | `10s` | Timeout de conexión inicial. |

## Pruebas y validación

```bash
gofmt -w .
go vet ./...
go test ./...
go build ./...
```

## Despliegue

Consulta [docs/DEPLOY_VERCEL.md](docs/DEPLOY_VERCEL.md). Resumen:

1. Crea una base PostgreSQL gestionada (Neon, Supabase o Vercel Marketplace).
2. Importa el repositorio en Vercel.
3. Define `DATABASE_URL` (pooled), `APP_ENV=production`, `COOKIE_SECURE=true` y, opcionalmente, el administrador inicial.
4. Despliega. `vercel.json` fija el preset `go` y Vercel arranca `main.go` usando `$PORT`.
5. Verifica `/healthz`, el registro, el login y la persistencia tras un redeploy.

## Notas

- Las sesiones y los usuarios persisten en PostgreSQL; no hay estado en memoria.
- Solo se guarda el hash SHA-256 del token de sesión, nunca el token en claro.
- Las contraseñas usan PBKDF2-HMAC-SHA256 con 600.000 iteraciones.
- El endpoint `/profile` devuelve un view model sin `PasswordHash`.
- El catálogo de vehículos sigue siendo una vista de demostración; `POST /cars` y `GET /users` son placeholders existentes.
