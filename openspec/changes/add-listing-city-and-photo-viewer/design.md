# Design

## Context

La aplicación es un monolito Go con Gin, PostgreSQL mediante `pgx` sobre `database/sql`, templates HTML y CSS embebidos con `go:embed`, y despliegue en Vercel. El detalle público vive en `internal/httpapi/handlers/home.go` y `internal/webui/templates/detail.html`; las fotografías se sirven por `/photos/:id`. El formulario de publicación y edición está en `publishing.go` y `publish.html`. La búsqueda se normaliza en `internal/domain/search.go` y se ejecuta en `CarRepository.Search`. Las migraciones son una lista idempotente de sentencias aplicada en una transacción con advisory lock.

Actualmente:

- Las miniaturas del detalle son enlaces `<a target="_blank">` al endpoint de la fotografía, por lo que abandonan la publicación.
- `cars` no tiene ciudad; el formulario, los view models, los filtros y las tarjetas no la contemplan.
- La edición está limitada a diez minutos; las publicaciones más antiguas no se pueden modificar.
- No hay framework de frontend ni build de JavaScript; los templates se renderizan en el servidor.

Ver `proposal.md` para la motivación y `specs/` para el comportamiento exigido.

## Goals / Non-Goals

**Goals:**

- Incorporar la ciudad como atributo obligatorio de cada publicación nueva o editada, visible en detalle y tarjetas.
- Permitir filtrar por ciudad en catálogo y Ventas, combinándolo con los filtros existentes y conservándolo en la paginación.
- Sustituir la apertura de fotografías en pestaña nueva por un visor modal accesible dentro del detalle.
- Mantener la migración aditiva y la compatibilidad con publicaciones existentes.
- No agregar dependencias ni un pipeline de frontend.

**Non-Goals:**

- Autocompletado, catálogo de municipios, geocodificación, distancias ni mapas.
- Backfill de ciudades para publicaciones heredadas o una acción especial para completarlas.
- Filtro por ciudad en el panel de moderación o en la API administrativa.
- Zoom, paneo, descarga o comparación de imágenes dentro del visor.
- Visor en las tarjetas del catálogo, en la vista previa del formulario o en la pantalla de edición.
- Cambiar la ventana de edición de diez minutos.

## Decisions

### 1. La ciudad es un atributo de la publicación, no del perfil del vendedor

Se agrega `City` a `domain.Car`, `CarInput` y `CarDetails`. El formulario la envía como texto libre y `NormalizeCarInput` la normaliza con `strings.Fields` (colapsa espacios) y valida que no esté vacía y no supere 80 caracteres.

- **Por qué:** un vendedor puede publicar vehículos en ciudades distintas y la ubicación debe viajar con la publicación.
- **Alternativas descartadas:** ciudad en el perfil del vendedor (todas sus publicaciones compartirían ubicación), lista de ciudades (requiere mantener un dataset y no fue la opción elegida), ciudad opcional (no permitiría un filtro confiable).
- **Errores:** se agregan `ErrCityRequired` y `ErrCityTooLong` para que el adaptador HTTP los mapee a `422` y a mensajes junto al campo.

### 2. Columna `cars.city` aditiva con valor vacío para datos heredados

La migración agrega `city TEXT NOT NULL DEFAULT ''`. Las filas existentes quedan con cadena vacía; no se rellena con una ciudad inventada. Las consultas usan `COALESCE(c.city, '')` por robustez y los view models traducen la cadena vacía a “Ciudad no especificada”.

- **Por qué:** evita un `NOT NULL` sin valor real, no bloquea el despliegue y mantiene legibles las publicaciones antiguas.
- **Alternativa descartada:** backfill con “No especificada” (contamina el dato y aparece como ciudad filtrable); columna nullable (complica los escaneos y no aporta nada frente a `DEFAULT ''`).
- **Validación de aplicación:** la obligatoriedad se exige al crear y al editar; la base de datos no impone `CHECK (city <> '')` para no romper filas heredadas.

### 3. Filtro de ciudad normalizado y parametrizado

`SearchQuery` y `SearchFilters` incorporan `City`. La normalización aplica `strings.Fields` y compara con `LOWER(c.city) = LOWER($n)`; el valor viaja siempre como parámetro. `HasFilters` incluye la ciudad y los helpers de paginación la conservan en la URL.

- **Por qué:** una comparación exacta sin distinguir mayúsculas evita coincidencias accidentales por subcadenas y mantiene el índice usable.
- **Limitación:** la comparación es sensible a tildes (`Bogotá` no coincide con `Bogota`). Se documenta; una canonicalización con `unaccent` requeriría una extensión que no está garantizada en todos los proveedores. La ciudad se muestra tal como la escribió el vendedor.
- **Alternativas descartadas:** `ILIKE '%texto%'` (sobrecoincidencias y peor uso de índices), `unaccent` (dependencia de extensión), filtro en memoria (incumple la búsqueda en backend).

### 4. Visor modal con mejora progresiva

El detalle renderiza la fotografía principal y las miniaturas como enlaces a `/photos/:id` sin `target="_blank"`. Un script propio intercepta el clic, evita la navegación y abre un contenedor modal (`role="dialog"`, `aria-modal="true"`) con la imagen, contador, botones anterior/siguiente y cerrar.

- **Por qué:** sin JavaScript, el enlace sigue abriendo la imagen en la misma pestaña; con JavaScript, la experiencia es un visor sin salir del detalle. No se depende de una librería ni de un build.
- **Estado:** el visor se renderiza una vez con la lista de URLs y textos alternativos; el índice actual se guarda en una variable del closure. Se abre desde la foto principal (índice 0) o desde una miniatura (índice correspondiente).
- **Interacción:** flechas izquierda/derecha y botones cambian de foto; `Esc`, el botón de cerrar y el clic en el fondo cierran. Con una sola foto se ocultan los controles de navegación.
- **Accesibilidad:** se guarda el elemento que abrió el visor, se mueve el foco al botón de cerrar o al contenedor, se encierra `Tab` dentro del diálogo, se usa `aria-live="polite"` para el contador y se restaura el foco al cerrar. Se bloquea el scroll del `body` mientras está abierto.
- **CSS:** fondo semitransparente, imagen centrada con `max-height`, controles con foco visible, diseño responsive y respeto a `prefers-reduced-motion`.
- **Alternativa descartada:** `<dialog>` nativo (comportamiento variable entre navegadores y menos control sobre el foco y el fondo); carrusel inline sin modal (no amplía la imagen y no cumple la comodidad solicitada).

### 5. Propagación por capas

- **Dominio:** `Car.City`, `CarInput.City`, `CarDetails.City`, `SearchFilters.City`, `SearchQuery.City`; validación y normalización.
- **Servicios:** `CreateCarInput.City` y `UpdateCarInput.City` se copian a `CarDetails` y al `domain.Car` actualizado; `NormalizeSearchFilters` incorpora la ciudad.
- **Persistencia:** `carSelectColumns` agrega `COALESCE(c.city, '')`; `scanCar` la lee; `CreateWithPhotos` y `UpdateWithPhotos` la escriben; `Search` agrega la condición de ciudad.
- **HTTP:** `views.CarForm`, `views.CarCard`, `views.CarView`; `readCarForm` lee `city`; los handlers de catálogo y Ventas leen `?city=`; los helpers de paginación la incluyen.
- **Templates:** input de ciudad en `publish.html`; campo de filtro en `catalog.html` y `sales.html`; ciudad en `partials.gohtml` (tarjeta), `detail.html` y `sales.html` (tabla de publicaciones propias); visor en `detail.html`.

### 6. Compatibilidad y migración

La migración es una sentencia aditiva dentro del arreglo existente. El código anterior ignora la columna nueva, por lo que el rollback inmediato consiste en desplegar la versión previa. No se eliminan ni renombran columnas.

- **Publicaciones heredadas:** se muestran como “Ciudad no especificada” y no aparecen al filtrar por ciudad. Si su propietario las edita dentro del plazo, el formulario exige ciudad. Fuera del plazo no hay una acción especial, conforme a la decisión de producto.
- **Datos nuevos:** toda publicación creada o editada después del despliegue requiere ciudad.

## Risks / Trade-offs

- **Ciudades escritas de forma inconsistente** → normalización de espacios y comparación sin distinguir mayúsculas; se documenta que el texto debe coincidir en tildes. Una evolución futura puede introducir un catálogo canónico.
- **Publicaciones heredadas sin ciudad** → se muestran con una etiqueta honesta y quedan fuera del filtro; no se inventan datos ni se reabre la edición.
- **Filtro exacto y sensible a tildes** → se acepta para evitar una extensión de PostgreSQL; el mensaje de estado vacío invita a limpiar filtros.
- **JavaScript deshabilitado** → los enlaces sin `target="_blank"` mantienen la navegación en la misma pestaña; el visor es una mejora progresiva.
- **Foco y teclado en el modal** → se implementa un ciclo de foco explícito y se prueba con teclado; se restauran el foco y el scroll al cerrar.
- **Crecimiento del template de detalle** → el visor reutiliza las URLs ya renderizadas y no agrega consultas ni endpoints.
- **Migración en producción** → la columna tiene `DEFAULT ''` y es aditiva; se aplica en la misma transacción con advisory lock y se verifica ejecutándola dos veces en `TEST_DATABASE_URL`.

## Migration Plan

1. Ejecutar `gofmt -w .`, `go vet ./...`, `go test ./...` y `go build ./...` antes del cambio.
2. Desplegar la migración aditiva `ALTER TABLE cars ADD COLUMN IF NOT EXISTS city TEXT NOT NULL DEFAULT ''`.
3. Desplegar el código que lee y escribe `city`, muestra la ciudad y agrega el filtro.
4. Verificar en un entorno con datos existentes que las publicaciones heredadas se muestran como “Ciudad no especificada” y que las nuevas exigen ciudad.
5. Verificar el visor en escritorio y móvil, con teclado y con JavaScript deshabilitado.
6. Rollback: desplegar la versión anterior; la columna nueva se ignora y no se pierden datos.

## Open Questions

Ninguna. Las decisiones de obligatoriedad, texto libre, filtro y tratamiento de publicaciones heredadas quedaron resueltas con el usuario.
