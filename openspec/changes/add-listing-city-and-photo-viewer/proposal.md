# Proposal

## Why

Hoy las miniaturas del detalle abren la fotografía original en una pestaña nueva, lo que saca al visitante del catálogo y rompe la comparación entre imágenes. Además, las publicaciones no indican desde qué ciudad se vende el vehículo, de modo que compradores y vendedores no pueden filtrar por ubicación ni evaluar la distancia antes de contactar.

## What Changes

- **Visor de fotografías en el detalle:** al hacer clic en la fotografía principal o en una miniatura se abre un visor modal dentro de la misma página, con navegación anterior/siguiente, contador, cierre y atajos de teclado. Se elimina la apertura en pestaña nueva.
- **Ciudad de la publicación:** el formulario de publicación y edición incorpora un campo de ciudad obligatorio, de texto libre, con validación en el servidor y límite de longitud.
- **Ciudad visible:** la ciudad se muestra en la página de detalle y en las tarjetas del catálogo y de Ventas.
- **Filtro por ciudad:** el catálogo público y el apartado Ventas permiten filtrar por ciudad, combinándolo con la búsqueda de texto, año y rango de precio; la paginación conserva el criterio.
- **Datos existentes:** las publicaciones creadas antes de la migración conservan su registro con ciudad vacía y se muestran como “Ciudad no especificada”. No se rellena ni se inventa una ubicación, y no se ofrece una acción especial para completarla fuera del plazo de edición.
- **Sin nuevas dependencias:** el visor se implementa con HTML, CSS y JavaScript del propio proyecto, manteniendo templates embebidos y el renderizado del servidor.

## Capabilities

### New Capabilities

Ninguna. El cambio extiende capacidades existentes.

### Modified Capabilities

- `vehicle-catalog`: el detalle incorpora la ciudad y un visor de fotografías en la misma página, sin pestañas nuevas; las tarjetas muestran la ciudad.
- `vehicle-publishing`: el formulario exige una ciudad válida al crear y editar, y la conserva al actualizar la publicación.
- `vehicle-search`: se agrega el filtro por ciudad, combinable con los criterios existentes y conservado en la paginación.

## Impact

- **Dominio:** `Car`, `CarInput`, `CarDetails` y `SearchFilters` incorporan la ciudad; se agregan errores de validación de ciudad y su normalización.
- **Servicios y puertos:** creación y edición propagan la ciudad; la búsqueda normaliza el filtro de ciudad.
- **Persistencia:** migración aditiva `cars.city`; las consultas de catálogo, Ventas y detalle leen la ciudad y el buscador agrega una condición parametrizada por ciudad.
- **HTTP y vistas:** formulario de publicación/edición, filtros de catálogo y Ventas, tarjetas, detalle y view models incluyen la ciudad; el detalle agrega el modal del visor.
- **UI:** nuevos estilos para el visor y para la ciudad en tarjetas y detalle, con foco visible, cierre con `Esc` y navegación accesible.
- **Documentación:** `README.md` y la documentación de arquitectura describen el campo de ciudad, el filtro y el visor.
- **Compatibilidad:** migración aditiva; no se eliminan columnas ni rutas. Las publicaciones heredadas sin ciudad siguen visibles.
