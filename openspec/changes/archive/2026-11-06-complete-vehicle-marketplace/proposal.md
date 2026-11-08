# Proposal

## Why

AutoMarket ya cuenta con autenticación, sesiones en servidor y un esqueleto de publicaciones, pero todavía no es una plataforma funcional de compraventa: el catálogo no tiene detalle ni fotos, la búsqueda no existe, los mensajes no están implementados y la moderación borra registros sin dejar evidencia. Además, el catálogo se llena con vehículos ficticios y la interfaz usa azul como color principal, en contradicción con los requisitos del producto.

Completar estas capacidades permite que un visitante consulte publicaciones reales sin iniciar sesión, que un vendedor publique y gestione sus vehículos, y que un administrador modere con trazabilidad, manteniendo la arquitectura por capas y el stack actuales.

## What Changes

- **Catálogo público real:** listado y detalle de vehículos activos con fotos, marca, modelo, año, precio en pesos colombianos (COP), descripción, vendedor y fecha; estados vacíos claros; eliminación de datos ficticios y de `seedCars`.
- **Publicación de vehículos:** formulario multipart con validación de campos, precio en COP y año; una o varias fotografías con vista previa y eliminación; visibilidad inmediata sin aprobación; protección contra doble envío.
- **Gestión del vendedor:** apartado **Ventas** con la misma fuente de datos que el catálogo, bloque de publicaciones propias, edición limitada a los primeros 10 minutos desde la creación y retirada propia sin límite de tiempo.
- **Fotografías en PostgreSQL:** almacenamiento `BYTEA` en tabla separada, validación del tipo real y tamaño, endpoint de streaming con caché y `ETag`; las consultas de catálogo solo leen metadatos.
- **Búsqueda y filtros:** búsqueda por marca/modelo/año/descripción, filtros combinables de año y rango de precio, limpieza de filtros, paginación y consultas parametrizadas en el backend.
- **Mensajería:** botón de interés con texto inicial, conversaciones asociadas a publicación/remitente/vendedor, bandeja, hilo, respuestas, permisos por participante, prevención de auto-mensajes y fechas de envío/lectura. Los visitantes anónimos deben autenticarse conservando el borrador.
- **Moderación con auditoría:** retirada administrativa con razón obligatoria, confirmación previa, borrado lógico mediante estado y registro inmutable de administrador, publicación, razón y fecha. La publicación retirada deja de aparecer en catálogo, Ventas y búsqueda.
- **Control de acceso backend:** autorización por rol y por propiedad en cada endpoint, no solo ocultando elementos de la interfaz.
- **Identidad visual:** nueva paleta sin azul como color principal ni de acento, diseño responsive y consistente en catálogo, detalle, publicación, Ventas, Mensajes y administración.
- **BREAKING (esquema interno):** `cars.price` pasa a un modelo numérico `price_cents` expresado en centavos de peso colombiano (COP); el borrado físico de publicaciones se reemplaza por estados `active`, `withdrawn` y `removed`; los vehículos pasan a requerir vendedor y al menos una foto.

## Capabilities

### New Capabilities

- `access-control`: roles visitante, vendedor y administrador; autorización por endpoint, propiedad de publicaciones y regla temporal de edición.
- `vehicle-catalog`: catálogo público de publicaciones activas, detalle del vehículo, información pública del vendedor y estados vacíos.
- `vehicle-publishing`: creación y edición de publicaciones por parte del vendedor, validación de datos y visibilidad inmediata.
- `vehicle-photos`: almacenamiento, validación y servicio de fotografías asociadas a cada publicación.
- `vehicle-search`: búsqueda de texto, filtros combinables, paginación y limpieza de criterios ejecutados en el backend.
- `seller-sales`: apartado Ventas con la misma fuente que el catálogo, publicaciones propias, edición temporal y retirada.
- `messaging`: conversaciones entre interesados y vendedores, bandeja, respuestas, permisos y conservación del borrador.
- `content-moderation`: retirada administrativa con razón, confirmación, borrado lógico y registro de auditoría.
- `visual-identity`: paleta sobria sin azul, jerarquía visual, accesibilidad, responsive y consistencia entre pantallas.

### Modified Capabilities

Ninguna. `openspec/specs/` está vacío, por lo que todas las capacidades son nuevas.

## Impact

- **Código afectado:** `internal/domain` (entidades y validaciones), `internal/service` (casos de uso y puertos), `internal/repository/postgres` (repositorios, búsqueda y migraciones), `internal/httpapi` (handlers, contratos y router), `internal/views`, `internal/webui` (templates y CSS) e `internal/app` (composition root y eliminación del seeding).
- **Base de datos:** nuevas tablas `car_photos`, `conversations`, `messages` y `moderation_logs`; alteraciones aditivas en `cars` (`status`, `price_cents`, `client_token`, `edited_at`) y backfill controlado del precio.
- **Endpoints:** catálogo y detalle públicos; publicación, edición, retirada y Ventas para vendedor; mensajería para participantes; moderación y auditoría para administrador; endpoint público de fotos.
- **Dependencias:** sin nuevas tecnologías; se mantienen Go, Gin, PostgreSQL, templates embebidos y el despliegue en Vercel. Se evalúa únicamente `image` de la librería estándar para validar JPEG/PNG.
- **Datos:** se eliminan los 6 vehículos demo sembrados por la aplicación mediante una limpieza idempotente y controlada.
- **Seguridad:** validación de archivos por contenido, límites de tamaño, consultas parametrizadas, autorización de propiedad y protección contra doble envío.
