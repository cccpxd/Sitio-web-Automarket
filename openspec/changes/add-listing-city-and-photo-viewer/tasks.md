# Tasks

## 1. Dominio y búsqueda

- [ ] 1.1 Agregar `City` a `domain.Car`, `domain.CarInput` y `domain.CarDetails`, normalizarla con espacios colapsados y validarla como obligatoria con máximo 80 caracteres; verificar con tests de ciudad válida, vacía y demasiado larga en `internal/domain/marketplace_test.go`.
- [ ] 1.2 Agregar `ErrCityRequired` y `ErrCityTooLong` en `internal/domain/errors.go` y usarlos desde `NormalizeCarInput`; verificar que los tests comparan los errores con `errors.Is`.
- [ ] 1.3 Agregar `City` a `domain.SearchQuery` y `domain.SearchFilters`, normalizarla y incluirla en `HasFilters`; verificar con tests de ciudad con espacios, ciudad vacía y combinación con año y precio.

## 2. Persistencia y migración

- [ ] 2.1 Agregar `ALTER TABLE cars ADD COLUMN IF NOT EXISTS city TEXT NOT NULL DEFAULT ''` a `internal/repository/postgres/migrate.go`, y leer `COALESCE(c.city, '')` en `carSelectColumns` y `scanCar`; verificar con la migración ejecutada dos veces y una consulta de lectura en `postgres_integration_test.go`.
- [ ] 2.2 Escribir `city` en `CreateWithPhotos` y `UpdateWithPhotos`; verificar con tests de integración que una publicación creada y editada conserva y actualiza la ciudad.
- [ ] 2.3 Agregar la condición de ciudad a `CarRepository.Search` usando un parámetro `LOWER(c.city) = LOWER($n)`; verificar con tests de integración de coincidencia sin mayúsculas, ciudad sin resultados, publicación con ciudad vacía y entrada con caracteres de inyección.
- [ ] 2.4 Extender `assertSchema` de `postgres_integration_test.go` para comprobar que `cars.city` existe con `NOT NULL` y valor por defecto vacío; verificar que la suite pasa con `TEST_DATABASE_URL`.

## 3. Servicios y puertos

- [ ] 3.1 Agregar `City` a `service.CreateCarInput` y `service.UpdateCarInput`, propagarla en `PublishingService.Create` y `PublishingService.Update`, y conservarla en el coche actualizado; verificar con tests de servicio de creación y edición con ciudad.
- [ ] 3.2 Actualizar los fakes de `internal/service/marketplace_test.go` para reflejar la ciudad y probar que la búsqueda del catálogo la normaliza y la pasa al store; verificar con un test que el filtro de ciudad llega a `SearchFilters.City`.

## 4. Formulario de publicación y edición

- [ ] 4.1 Agregar `City` a `views.CarForm`, leer `city` en `readCarForm`, enviarla en `Create` y `Update` y precargarla en `EditForm` y `editPage`; verificar con tests HTTP que el formulario de edición muestra la ciudad actual y que el POST la envía al servicio.
- [ ] 4.2 Agregar el campo de ciudad a `publish.html` con etiqueta, `maxlength="80"`, `required` y el error de campo correspondiente; verificar con un test de render que el HTML contiene el campo y el mensaje de error.
- [ ] 4.3 Mapear `ErrCityRequired` y `ErrCityTooLong` a `422`, a un mensaje junto al campo y a un mensaje general en `publishErrorMessage` y `publishStatus`; verificar con tests HTTP de ciudad vacía y demasiado larga.
- [ ] 4.4 Conservar la ciudad introducida cuando el formulario se rechaza por otro campo; verificar con un test HTTP que la respuesta mantiene el valor de ciudad junto con los demás datos válidos.
- [ ] 4.5 Actualizar `README.md` con el campo de ciudad obligatorio en el flujo de publicación; verificar que el texto coincide con el formulario implementado.

## 5. Catálogo, tarjetas, detalle y filtros

- [ ] 5.1 Agregar `City` a `views.CarCard`, `views.CarView` y `views.MyListing`, y construirla en `NewCarCard`, `NewCarView` y `NewMyListing` con “Ciudad no especificada” cuando esté vacía; verificar con tests unitarios de view models para ciudad presente y ausente.
- [ ] 5.2 Mostrar la ciudad en `partials.gohtml` (tarjeta), `detail.html` (metadatos) y `sales.html` (tabla de publicaciones propias); verificar con tests de render que la ciudad aparece y que una publicación heredada muestra “Ciudad no especificada”.
- [ ] 5.3 Leer `?city=` en los handlers de catálogo y Ventas, incluirlo en los helpers de paginación y en `HasFilters`; verificar con tests HTTP de filtro por ciudad y de paginación que conserva el criterio.
- [ ] 5.4 Agregar el input de ciudad a `catalog.html` y `sales.html` y ajustar la cuadrícula de filtros en `style.css` para el nuevo campo en escritorio y móvil; verificar con tests de render y revisión visual responsive.
- [ ] 5.5 Actualizar `docs/ARCHITECTURE.md` con la columna `cars.city`, el filtro de ciudad y el flujo de búsqueda; verificar que la descripción coincide con las rutas y consultas implementadas.

## 6. Visor de fotografías

- [ ] 6.1 Sustituir los enlaces `target="_blank"` de `detail.html` por enlaces a `/photos/:id` sin destino nuevo, con atributos de datos para índice y URL, y agregar el contenedor del visor con `role="dialog"`, `aria-modal`, contador y controles; verificar con un test de render que no existe `target="_blank"` y que el modal está presente.
- [ ] 6.2 Implementar en `detail.html` el JavaScript del visor: abrir desde la foto principal o una miniatura, navegación anterior/siguiente, cierre con botón, `Esc` y fondo, contador con `aria-live`, bloqueo de scroll, ciclo de foco y restauración del foco; verificar manualmente en navegador con teclado y con una y varias fotografías.
- [ ] 6.3 Agregar estilos del visor en `style.css` (fondo, imagen, controles, foco visible, responsive y `prefers-reduced-motion`); verificar visualmente en escritorio y móvil que la imagen se amplía y los controles son usables.
- [ ] 6.4 Verificar la mejora progresiva sin JavaScript: los enlaces abren la fotografía en la misma pestaña y nunca en una ventana nueva; registrar el resultado de la prueba manual.
- [ ] 6.5 Documentar el visor y su navegación por teclado en `README.md`; verificar que las instrucciones coinciden con el comportamiento implementado.

## 7. Integración y aceptación

- [ ] 7.1 Ejecutar `gofmt -w .`, `go vet ./...`, `go test ./...` y `go build ./...`; verificar que todos terminan sin errores.
- [ ] 7.2 Ejecutar `openspec validate add-listing-city-and-photo-viewer --strict`; verificar que no hay errores ni advertencias.
- [ ] 7.3 Recorrer manualmente el flujo completo: publicar con ciudad, verla en tarjeta y detalle, filtrar por ciudad en catálogo y Ventas, comprobar la paginación con el filtro y abrir el visor en escritorio y móvil.
- [ ] 7.4 Verificar con una publicación heredada sin ciudad que se muestra “Ciudad no especificada”, no aparece al filtrar por ciudad y no se inventa una ubicación.

## Workflow follow-up

- Archivar el cambio después de completar la implementación y la revisión del proyecto.
- Verificar que las specs principales quedaron actualizadas con la ciudad y el visor.
