# ADR-0004: Control de acceso por rol con middleware de Gin

- **Estado:** Aceptado
- **Fecha:** 2026-09-27
- **Autores:** Cristian Camilo Colorado; Juan Jose Patiño; Juan Manuel Ramos

## Contexto

El sistema tiene tres roles: Visitante, Vendedor y Administrador. Cada rol puede hacer cosas distintas (autorización, parte de AAA). Los permisos deben revisarse en un solo lugar y no repetirse en cada función.

## Decisión

Usamos middlewares de Gin que se ponen en las rutas.

- El middleware de autenticación (ADR-0002) identifica al usuario y su rol.
- Un middleware `RequerirRol(roles ...)` deja pasar solo a los roles permitidos.
- Sin sesión válida, la respuesta es 401. Con sesión pero sin permiso, la respuesta es 403.
- El Visitante es quien no tiene sesión. Solo puede usar las rutas públicas.
- Las rutas se agrupan por rol con `router.Group(...)`.

## Alternativas consideradas

1. **Un `if` de rol dentro de cada handler.** Se repite y es fácil olvidarlo en una ruta. Descartada.
2. **Tabla de permisos (acción → roles) consultada por un middleware.** Es más ordenada, pero es más trabajo para el tamaño del proyecto. Puede ser una mejora futura.

## Consecuencias

- **A favor:** los permisos se ven de un vistazo en la definición de las rutas. El código de los handlers queda limpio.
- **En contra:** una ruta sin middleware queda abierta a todos. Hay que revisar esto en cada pull request.
