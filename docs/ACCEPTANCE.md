# Verificación de aceptación del marketplace

Fecha de verificación: 2026-11-06.

| # | Escenario | Evidencia |
|---:|---|---|
| 1 | Visitante consulta el catálogo | `internal/httpapi/router/router_test.go`; ruta pública `/cars` sin sesión |
| 2 | Búsqueda `Mazda 2013` | `domain.NormalizeSearchFilters` y `CarRepository.Search` parametrizado |
| 3 | Vendedor publica con fotografías | `PublishingService.Create`, `CreateWithPhotos` transaccional y formulario multipart |
| 4 | Otro vendedor ve la publicación en Ventas | `/seller/ventas` usa `CatalogService.Search`, igual que `/cars` |
| 5 | Catálogo y Ventas comparten activos | ambos handlers consumen el mismo puerto de catálogo |
| 6 | Mensaje y respuesta | `MessagingService`, rutas `/cars/:id/contacto` y `/mensajes/:id` |
| 7 | Tercero no lee conversación | autorización de participante en servicio y repositorio |
| 8 | Administrador retira con razón | formulario de confirmación y `ModerationService.Remove` |
| 9 | Retirada desaparece de catálogo/búsqueda | filtros `status = 'active'` en `CarRepository.Search` y detalle activo |
| 10 | Auditoría conserva razón | `moderation_logs` y `ModerationRepository.RemoveWithAudit` |
| 11 | Catálogo vacío honesto | template `empty-catalog`; no existe seed de vehículos en `internal/app` |
| 12 | Datos e imágenes inválidas se rechazan | validación de dominio, `ImageProcessor` y límites multipart |
| 13 | Permisos directos | middleware, servicios y consultas de propiedad/participación |

## Comandos ejecutados

```text
go test ./...
go vet ./...
go build ./...
openspec validate complete-vehicle-marketplace --strict
```

La prueba de PostgreSQL se ejecuta automáticamente cuando se define `TEST_DATABASE_URL`; sin esa variable se omite de forma explícita. El test de integración cubre migración idempotente, tablas nuevas y columnas no nulas.

La vista previa permite quitar archivos seleccionados antes del envío mediante `DataTransfer`; el endpoint vuelve a validar tipo real, dimensiones, tamaño y cantidad, por lo que la interfaz no es una frontera de seguridad.
