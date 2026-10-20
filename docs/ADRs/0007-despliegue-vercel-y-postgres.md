# ADR-0007: Despliegue en Vercel con PostgreSQL y assets embebidos

- **Estado:** Aceptado
- **Fecha:** 2026-10-19
- **Reemplaza:** ADR-0001 (persistencia en memoria/JSON)
- **Ajusta:** ADR-0002 (sesión en cookie HttpOnly) y ADR-0005 (logs a stdout)

## Contexto

AutoMarket debía desplegarse en Vercel. La versión inicial usaba Gin con puerto fijo `:8081`, templates y estáticos leídos del filesystem, y un repositorio en memoria con un administrador hardcodeado. Ese diseño funciona en local, pero en Vercel:

- el puerto lo define `PORT`;
- el filesystem es efímero y no escribible de forma persistente;
- cada invocación puede usar una instancia distinta, por lo que la memoria no se comparte;
- los secretos no pueden vivir en el repositorio.

El runtime Go de Vercel está disponible como Framework Preset (Beta), detecta `main.go` y `go.mod`, y soporta servidores `net/http` y frameworks como Gin.

## Decisión

1. Desplegar el servidor Go/Gin con el preset `go` de Vercel. `main.go` escucha en `:$PORT`.
2. Persistir usuarios y sesiones en PostgreSQL mediante `pgx` sobre `database/sql`.
3. Aplicar el esquema de forma idempotente al arrancar, con transacción y advisory lock.
4. Embeber templates y CSS con `go:embed` para que el binario sea autocontenido.
5. Leer toda la configuración de variables de entorno y validarla en `internal/config`.
6. Guardar solo el fingerprint SHA-256 del token de sesión; la cookie es `HttpOnly`, `Secure` y `SameSite=Lax`.
7. Crear el administrador inicial desde `ADMIN_EMAIL`/`ADMIN_PASSWORD`, de forma idempotente y sin promover cuentas existentes. Si el email pertenece a un no-admin, el arranque falla.
8. Escribir logs estructurados a stdout.

## Alternativas consideradas

1. **Funciones `api/*.go`:** cada archivo es una función independiente. Para servir todas las rutas de una app Gin haría falta un adaptador y configuración extra. Descartada.
2. **Frontend estático + API externa:** separa despliegues y añade CORS, dominio y observabilidad duplicada. Válida, pero innecesaria para este tamaño. Descartada.
3. **SQLite/JSON persistente:** el filesystem de Vercel no es persistente. Descartada.
4. **Mantener memoria:** pierde usuarios y sesiones entre invocaciones. Descartada.
5. **Render/Fly/Railway para el backend:** buena opción para procesos de larga duración, pero Vercel ya ejecuta el servidor Go. Descartada para evitar dos proveedores.

## Consecuencias

- **A favor:** un solo despliegue, datos persistentes, secretos fuera del código, assets autocontenidos, capas testeables sin PostgreSQL/Gin y separación estructural explícita.
- **En contra:** el preset Go está en Beta; hay cold starts; PBKDF2 consume CPU; la concurrencia obliga a cuidar el pool de conexiones; se depende de un PostgreSQL externo; la región de la función debe acercarse a la de la base de datos.
- **Operación:** la migración y el bootstrap corren en cada arranque, pero son idempotentes. La limpieza de sesiones es oportunista en login; para alto volumen se recomienda un cron.
- **Seguridad pendiente:** CSRF explícito, rate limiting, verificación de email, recuperación de contraseña y MFA no forman parte de este cambio.
