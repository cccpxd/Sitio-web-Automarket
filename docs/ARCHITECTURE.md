# Arquitectura de AutoMarket

## Principio rector

El proyecto aplica separación estructural: **entrada**, **procesamiento** y **salida** viven en paquetes distintos. Las dependencias apuntan hacia el dominio; el dominio no conoce ningún framework.

```
Entrada HTTP          Procesamiento              Salida
httpapi/handlers  ->  service  ->  domain   <-  repository/postgres
httpapi/middleware        ^                       webui (templates/estáticos)
                          |
                    puertos (interfaces)
                          ^
                  security / platform
```

`internal/app` es el único lugar que conoce todas las implementaciones. Es el composition root.

## Capas

| Paquete | Responsabilidad | Puede importar | No debe importar |
| --- | --- | --- | --- |
| `internal/domain` | Entidades, errores de negocio, normalización y validación pura | Librería estándar | Gin, SQL, templates, `os` |
| `internal/service` | Casos de uso y puertos de persistencia/criptografía | `domain` | Gin, SQL, HTML, variables de entorno |
| `internal/repository/postgres` | Persistencia y migraciones | `domain`, `database/sql`, `pgx` | Gin, templates, servicios |
| `internal/security` | Adaptadores de PBKDF2, tokens e IDs | Librería estándar | Dominio, HTTP, base de datos |
| `internal/httpapi` | Contratos de entrada/salida y adaptadores Gin | `domain`, `service` | Implementaciones concretas de persistencia |
| `internal/httpapi/handlers` | Bind de formularios, mapeo de errores, respuesta HTML/JSON | `httpapi`, `views`, Gin | SQL, reglas de negocio |
| `internal/httpapi/middleware` | Autenticación, autorización, logs, headers | `httpapi`, `domain`, Gin | Repositorios concretos |
| `internal/httpapi/router` | Composición de rutas y middlewares | handlers, middleware | Lógica de negocio |
| `internal/views` | View models; nunca expone `PasswordHash` | `domain` | Gin, SQL |
| `internal/webui` | Templates y estáticos embebidos + renderer | Librería estándar | Servicios, Gin |
| `internal/config` | Lectura y validación de entorno | Librería estándar | Dominio, Gin |
| `internal/platform` | Servidor HTTP, logger, reloj | Librería estándar, `config` | Dominio |
| `internal/app` | Composition root e inyección de dependencias | Todos los adaptadores | — |

## Flujo de una petición

1. `router` selecciona el handler y los middlewares.
2. `middleware` valida la cookie de sesión llamando al puerto `SessionValidator`.
3. `handlers` hace bind del formulario y construye un input de aplicación.
4. `service` normaliza/valida, aplica reglas y llama a los puertos `UserStore`/`SessionStore`.
5. `repository/postgres` ejecuta SQL y traduce errores de infraestructura a errores de dominio.
6. `handlers` mapea el error de dominio a HTTP y renderiza un view model con `webui.Renderer`.
7. `middleware` registra una línea estructurada con método, ruta, estado y duración.

## Inversión de dependencias

Los puertos se declaran en el paquete consumidor:

- `service.UserStore`, `service.SessionStore`, `service.PasswordHasher`, `service.TokenIssuer`, `service.IDGenerator`, `service.Clock`.
- `httpapi.Renderer`, `httpapi.AuthService`, `httpapi.SessionValidator`.

Los adaptadores (`repository/postgres`, `security`, `webui`) implementan esos contratos sin que el servicio los importe. Esto permite probar los casos de uso con dobles en memoria sin levantar PostgreSQL ni Gin.

## Reglas de diseño aplicadas

- `main` solo carga configuración, construye la aplicación y arranca el servidor.
- No hay variables globales de negocio ni singletons de persistencia.
- Cada función tiene una responsabilidad: normalizar, validar, mapear errores, renderizar, registrar, etc.
- Las funciones reciben dependencias y datos por parámetro y devuelven resultados.
- La lógica de negocio no depende de Gin, HTML, PostgreSQL ni del sistema de archivos.
- Las respuestas JSON usan `views.UserView`; `domain.User` nunca se serializa directamente.
- Los templates y estáticos se embeben en el binario para que el despliegue no dependa del filesystem.
- Los tokens de sesión se persisten solo como fingerprint SHA-256.

## Persistencia

PostgreSQL es la única fuente de verdad en producción. El esquema se aplica al arrancar con `CREATE TABLE IF NOT EXISTS` dentro de una transacción y un advisory lock, lo que evita carreras entre cold starts.

Tablas:

- `users`: cuenta, hash de contraseña y rol.
- `sessions`: fingerprint del token, usuario, creación y expiración.

No hay almacenamiento en memoria en el camino de producción.

## Extensión

Para añadir un módulo (por ejemplo, vehículos):

1. Define la entidad y las reglas en `internal/domain`.
2. Declara el puerto en `internal/service` y el caso de uso.
3. Implementa el adaptador en `internal/repository/postgres`.
4. Añade handlers y view models.
5. Registra las rutas en `internal/httpapi/router`.
6. Conecta las piezas en `internal/app`.

No se debe importar Gin desde `domain` o `service`, ni SQL desde `httpapi`.
