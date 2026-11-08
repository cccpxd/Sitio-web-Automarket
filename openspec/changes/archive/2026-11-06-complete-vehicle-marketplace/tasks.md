# Tasks

## 1. Preparación y línea base

- [x] 1.1 Ejecutar `go test ./...` y registrar el estado inicial; verificar que los tests existentes pasan o se omiten por falta de `TEST_DATABASE_URL`.
- [x] 1.2 Eliminar `seedCars` y su llamada en `internal/app/app.go`; verificar que la aplicación arranca sin insertar vehículos y que el catálogo queda vacío.
- [x] 1.3 Quitar de la landing las métricas y beneficios ficticios (`240+`, `4.9/5`, financiamiento, verificación, entrega); verificar que la página no muestra datos inventados.
- [x] 1.4 Agregar `golang.org/x/image` para el escalado de fotos y ejecutar `go mod tidy`; verificar que `go build ./...` compila.

## 2. Esquema y migraciones

- [x] 2.1 Agregar columnas aditivas a `cars` (`status`, `price_cents`, `client_token`, `edited_at`) con restricciones e índices; verificar migración ejecutándola dos veces en `TEST_DATABASE_URL`.
- [x] 2.2 Crear `car_photos`, `conversations`, `messages` y `moderation_logs` con claves foráneas, `CHECK` e índices; verificar que las tablas existen y que la segunda migración no falla.
- [x] 2.3 Implementar el backfill de `price_cents` desde `price` en pesos colombianos (COP) y la limpieza idempotente de los seis vehículos demo; verificar con consultas que no quedan filas demo ni precios nulos.
- [x] 2.4 Imponer `NOT NULL` en `seller_id` y `price_cents` tras la limpieza; verificar que la migración se aplica sin error y que los constraints aparecen en el esquema.
- [x] 2.5 Extender `postgres_integration_test.go` para cubrir el esquema nuevo, las restricciones y la idempotencia; verificar que el test pasa con `TEST_DATABASE_URL`.

## 3. Dominio

- [x] 3.1 Modelar `Car` con `Status` y `PriceCents` en COP, y agregar validaciones de estado, año, precio entero y descripción; verificar con tests unitarios de valores válidos e inválidos.
- [x] 3.2 Implementar la validación de fotos (tipo real, tamaño, cantidad y dimensiones); verificar con tests que aceptan JPEG/PNG y rechazan archivos falsos o excesivos.
- [x] 3.3 Definir `Conversation`, `Message` y `ModerationEntry` con validaciones de contenido, razón y auto-mensaje; verificar con tests unitarios de los límites.
- [x] 3.4 Definir `SearchFilters` y su normalización (texto, año, rango de precio, página); verificar con tests de combinaciones y valores inválidos.

## 4. Puertos y servicios

- [x] 4.1 Implementar el servicio de catálogo con listado activo, detalle, paginación y búsqueda; verificar con fakes que solo devuelve publicaciones activas y aplica todos los filtros.
- [x] 4.2 Implementar el servicio de publicación con creación, edición, ventana de diez minutos, retirada propia e idempotencia; verificar con tests de creación, doble envío, edición dentro/fuera del plazo y retirada.
- [x] 4.3 Implementar el servicio de fotos con normalización a JPEG, límites y metadatos; verificar con tests que el binario se reduce y que las imágenes inválidas no se guardan.
- [x] 4.4 Implementar el servicio de mensajería con inicio, reutilización de conversación, respuesta, lectura y permisos; verificar con tests de participante, tercero y auto-mensaje.
- [x] 4.5 Implementar el servicio de moderación con razón obligatoria, retirada lógica y auditoría; verificar con tests de razón vacía, retirada válida y conservación del registro.

## 5. Persistencia

- [x] 5.1 Implementar el repositorio de coches con filtros, estados, paginación, propiedad y corte de edición; verificar con tests de integración de búsqueda `Mazda 2013`, filtros combinados y exclusión de no activos.
- [x] 5.2 Implementar el repositorio de fotos separando consultas de metadatos y de contenido; verificar con tests que el listado no lee `data` y que el endpoint obtiene una sola fila.
- [x] 5.3 Implementar los repositorios de conversaciones y mensajes con alcance por participante; verificar con tests que un tercero no obtiene filas y que el administrador solo lee.
- [x] 5.4 Implementar el repositorio de moderación y la transacción de retirada + auditoría; verificar con tests que un fallo revierte ambos cambios y que el registro persiste.
- [x] 5.5 Implementar las transacciones de creación de coche + fotos y de conversación + primer mensaje; verificar con tests que no quedan registros parciales ante un fallo.

## 6. Catálogo y búsqueda

- [x] 6.1 Implementar el handler y template del catálogo con tarjetas, estado vacío y paginación; verificar con tests HTTP que un visitante sin sesión lo consulta y ve el mensaje “Todavía no hay vehículos publicados”.
- [x] 6.2 Implementar el detalle `/cars/:id` con fotos, datos, vendedor y fecha; verificar con tests de publicación activa, inexistente y retirada.
- [x] 6.3 Implementar el formulario de búsqueda y filtros combinables con limpieza, incluyendo precios en COP; verificar con tests de `Mazda 2013`, rango de precio, año y resultados vacíos.
- [x] 6.4 Implementar el endpoint de fotos con `ETag`, `Content-Length`, caché y acceso según estado; verificar con tests de descarga, `304` y denegación en publicación retirada.
- [x] 6.5 Implementar el formateo de precios en COP con separador de miles y sin decimales; verificar con tests unitarios de `45000000` → `$ 45.000.000` y de montos con y sin separadores.

## 7. Publicación y edición

- [x] 7.1 Implementar el formulario de publicación multipart con CSRF y validación de campos, incluido el precio entero en COP; verificar con tests de creación válida y rechazo por campos faltantes, año o precio inválidos.
- [x] 7.2 Implementar la vista previa, eliminación y cambio de fotos antes de guardar; verificar manualmente en navegador que la selección se refleja y que las fotos quitadas no se envían.
- [x] 7.3 Implementar la idempotencia con `client_token` y la redirección al resultado existente; verificar con un test de doble POST que solo existe una publicación.
- [x] 7.4 Implementar la edición de datos y fotos dentro de los diez minutos; verificar con tests de edición a los 9 minutos, rechazo a los 11, edición por otro vendedor y edición de una publicación retirada.
- [x] 7.5 Implementar la retirada por el vendedor con confirmación; verificar con tests que cambia a `withdrawn` y desaparece del catálogo y la búsqueda.
- [x] 7.6 Implementar la normalización de imágenes y los límites de tamaño, cantidad y dimensiones; verificar con tests de JPEG/PNG válidos, extensión falsa, archivo grande y exceso de fotos.

## 8. Ventas

- [x] 8.1 Implementar `/seller/ventas` con la misma consulta de publicaciones activas que el catálogo; verificar con un test que ambos listados contienen exactamente los mismos vehículos activos.
- [x] 8.2 Implementar el bloque “Mis publicaciones” con estados y acciones; verificar que el vendedor ve sus activas y retiradas, y no las retiradas de otros.
- [x] 8.3 Verificar que una publicación nueva aparece en Ventas y en el catálogo; cubrir con test de integración HTTP.
- [x] 8.4 Ocultar o deshabilitar la acción de editar fuera del plazo y mostrar el mensaje de expiración; verificar con test HTTP y con la comprobación del endpoint directo.

## 9. Mensajería

- [x] 9.1 Implementar el botón de interés y el texto inicial con descripción o referencia breve; verificar con tests de descripción corta y demasiado extensa.
- [x] 9.2 Implementar el inicio de conversación, la reutilización por `(car_id, buyer_id)` y el rechazo de auto-mensajes; verificar con tests de envío, reenvío y contacto a la propia publicación.
- [x] 9.3 Implementar la bandeja, el hilo, la respuesta y las fechas de envío/lectura; verificar con tests de conversación del vendedor y del comprador.
- [x] 9.4 Implementar los permisos de lectura y respuesta; verificar con tests que un tercero recibe error y que el administrador lee sin poder responder.
- [x] 9.5 Implementar la redirección a login/registro con `next` validado y la restauración del borrador; verificar con test HTTP del parámetro y prueba manual de `sessionStorage`.
- [x] 9.6 Implementar el aviso de publicación retirada en el hilo y bloquear nuevos contactos; verificar con tests de conversación histórica y de intento de contacto.

## 10. Moderación

- [x] 10.1 Implementar la pantalla de confirmación y el formulario con razón obligatoria; verificar con tests de retirada con razón y rechazo sin razón.
- [x] 10.2 Implementar la transacción de retirada administrativa y auditoría; verificar con test de integración que actualiza estado e inserta el registro.
- [x] 10.3 Verificar el ocultamiento en catálogo, detalle, búsqueda, Ventas y fotos; cubrir con tests HTTP.
- [x] 10.4 Implementar el registro de moderación en `/admin/moderacion`; verificar que muestra administrador, publicación, razón y fecha, y que un vendedor recibe error de permisos.
- [x] 10.5 Verificar que la retirada no borra la publicación ni sus conversaciones; cubrir con test de integración de integridad referencial.

## 11. Seguridad y errores

- [x] 11.1 Implementar el token CSRF de doble envío y aplicarlo a todos los formularios POST; verificar con tests de POST sin token y con token válido.
- [x] 11.2 Cambiar el logout a POST y conservar la compatibilidad necesaria; verificar que un GET no cierra la sesión y que el POST sí.
- [x] 11.3 Mapear errores de autenticación, permisos, conflicto, archivo demasiado grande y validación; verificar con tests HTTP de `401`, `403`, `409`, `413` y `422`.
- [x] 11.4 Validar el parámetro `next` como ruta local para evitar redirecciones abiertas; verificar con tests de URL externa y ruta interna.
- [x] 11.5 Revisar todas las consultas nuevas y probar entradas de inyección SQL en buscador y filtros; verificar que se tratan como texto y no alteran la consulta.

## 12. Identidad visual

- [x] 12.1 Reescribir las variables y estilos base con la paleta carbón/blanco/gris y acento terracota; verificar que no quedan colores azules como principal o acento.
- [x] 12.2 Extraer partials de encabezado, navegación, pie, flash y CSRF, y aplicarlos a todas las páginas; verificar que los tests de render de cada template pasan.
- [x] 12.3 Ajustar tarjetas, detalle, formularios, Ventas, Mensajes y administración a la nueva jerarquía visual; verificar visualmente en escritorio.
- [x] 12.4 Revisar responsive y accesibilidad en anchos de móvil, tableta y escritorio; verificar navegación por teclado, foco visible, etiquetas y contraste AA.
- [x] 12.5 Verificar que catálogo, detalle, publicación, Ventas, Mensajes y administración comparten encabezado, tipografía, botones y estados.

## 13. Documentación

- [x] 13.1 Actualizar `docs/ARCHITECTURE.md` con las nuevas tablas, rutas, servicios y regla de autorización; verificar que el documento coincide con el código implementado.
- [x] 13.2 Agregar ADRs para almacenamiento de fotos, borrado lógico y precio en COP; verificar que cada decisión incluye alternativas y consecuencias.
- [x] 13.3 Actualizar `README.md` y la documentación de despliegue con los nuevos flujos y límites de fotos; verificar que las instrucciones ejecutadas funcionan.

## 14. Integración y aceptación

- [x] 14.1 Ejecutar `go test ./...` y la suite de integración con `TEST_DATABASE_URL`; verificar que todo pasa.
- [x] 14.2 Recorrer manualmente los trece escenarios de aceptación del borrador; verificar cada uno y registrar el resultado.
- [x] 14.3 Arrancar contra una base de datos vacía y verificar catálogo vacío, estado informativo y ausencia de datos ficticios.
- [x] 14.4 Ejecutar `openspec validate complete-vehicle-marketplace --strict`; verificar que no hay errores ni advertencias.

## Workflow follow-up

- Archivar el cambio después de completar la implementación y la revisión del proyecto.
- Verificar el resultado archivado y que las specs principales quedaron actualizadas.

