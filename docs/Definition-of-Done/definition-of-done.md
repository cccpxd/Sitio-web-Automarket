# Definition of Done (DoD) — AutoMarket

Una historia de usuario está **terminada** solo si cumple **todos** estos puntos.

## Código
- [ ] Compila y corre sin errores.
- [ ] Cumple todos los criterios de aceptación de la historia.
- [ ] Usa solo la librería estándar de Go y Gin. No hay librerías prohibidas.
- [ ] No hay variables globales.
- [ ] Respeta las capas: handler, servicio y repositorio (ADR-0006).
- [ ] El código está formateado (`gofmt`) y pasa `go vet`.

## Pruebas
- [ ] La lógica de negocio tiene pruebas y todas pasan (`go test ./...`).
- [ ] Se probó a mano el flujo completo de la historia.

## Seguridad (AAA)
- [ ] Las rutas nuevas tienen el middleware de autenticación y de rol que les corresponde.
- [ ] Se probó que un rol sin permiso recibe 403, y sin sesión recibe 401.
- [ ] No se imprimen ni se guardan en logs contraseñas, hashes ni tokens.
- [ ] Las acciones importantes quedan registradas con `slog` (ADR-0005).

## Revisión e integración
- [ ] Se trabajó en una rama propia.
- [ ] Hay un pull request revisado y aprobado por al menos otro compañero.
- [ ] Está integrado en la rama principal y esta sigue compilando.

## Documentación y tablero
- [ ] Si se tomó una decisión de arquitectura nueva, hay un ADR.
- [ ] La tarjeta del kanban pasó a **Hecho**.
