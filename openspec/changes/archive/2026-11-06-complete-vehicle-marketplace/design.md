# Design

## Context

El proyecto es una aplicación Go 1.25 con Gin, PostgreSQL mediante `pgx` sobre `database/sql`, templates HTML y CSS embebidos con `go:embed`, y despliegue en Vercel. La arquitectura actual separa entrada HTTP, servicios, dominio puro y adaptadores de persistencia; los puertos se declaran en el paquete consumidor y `internal/app` es el único composition root.

El estado de partida condiciona el diseño:

- `cars` existe con `price TEXT`, `seller_id` opcional, `reported` y sin estado de publicación.
- El borrado administrativo es físico y no deja auditoría.
- No hay tablas de fotos, conversaciones, mensajes ni moderación.
- El catálogo lista todos los registros, sin búsqueda ni paginación.
- La aplicación siembra seis vehículos ficticios al arrancar.
- Los templates son archivos independientes, sin partials compartidos, y el CSS usa azul como color principal.
- El runtime de Vercel es de solo lectura y sin sistema de archivos persistente, por lo que templates y estáticos se embeben y los binarios deben vivir en PostgreSQL.

Ver `proposal.md` para la motivación y `specs/` para el comportamiento exigido.

## Goals / Non-Goals

**Goals:**

- Completar catálogo, publicación, búsqueda, mensajería y moderación respetando la arquitectura por capas existente.
- Mantener PostgreSQL como única fuente de verdad, incluidas las fotografías.
- Preservar la integridad de conversaciones y auditoría mediante borrado lógico.
- Aplicar autorización por rol y propiedad en el backend.
- Dejar una migración aditiva, verificable y reversible a nivel de aplicación.
- Sustituir la identidad visual azul por una paleta sobria y accesible.

**Non-Goals:**

- No se introduce un framework de frontend ni una API JSON separada; se mantiene el renderizado del servidor con formularios HTML.
- No se agrega un rol de comprador ni registro de visitantes; los interesados usan las cuentas existentes.
- No se implementa edición después de diez minutos, reactivación de publicaciones retiradas ni edición por administradores.
- No se migran imágenes a un CDN ni se genera una galería con transformaciones avanzadas.
- No se reemplaza el sistema de sesiones, el hashing PBKDF2 ni el despliegue en Vercel.

## Decisions

### 1. Fotografías como `BYTEA` en una tabla separada

Se almacena el binario normalizado en `car_photos`, con una fila por fotografía y metadatos (`content_type`, `byte_size`, `sha256`, dimensiones, posición). Las consultas de catálogo leen solo metadatos; el contenido se sirve por un endpoint individual.

- **Por qué:** mantiene la copia de seguridad y la transacción en PostgreSQL, sin proveedor externo ni credenciales adicionales.
- **Alternativas descartadas:** Large Objects (`lo_*`) complican el pool de conexiones y el runtime serverless; el almacenamiento externo agrega infraestructura, coste y limpieza de huérfanos para un volumen pequeño.
- **Normalización:** al aceptar una imagen se decodifica, se valida y se re-codifica a JPEG con un máximo de 1600 px en el lado mayor y calidad 85. Esto acota el tamaño real, evita enviar originales de varios megabytes y mantiene las respuestas dentro de los límites de Vercel. Se usa `golang.org/x/image/draw` para el escalado; es la única dependencia nueva y no cambia el stack.
- **Consecuencias:** la base de datos crece con las imágenes; se mitiga con límites (máximo 5 MB por archivo, 6 fotos por publicación) y con caché HTTP inmutable.

### 2. Estados de publicación y borrado lógico

`cars.status` admite `active`, `withdrawn` y `removed`. La retirada del vendedor usa `withdrawn`; la del administrador usa `removed` y exige razón. Ninguna elimina la fila.

- **Por qué:** conserva mensajes, fotos y auditoría; permite que el vendedor vea sus publicaciones retiradas y que el administrador mantenga evidencia.
- **Alternativa descartada:** borrado físico con `ON DELETE CASCADE`; rompería conversaciones y la trazabilidad exigida.
- **Regla de visibilidad:** solo `active` aparece en catálogo, búsqueda y Ventas pública; `withdrawn` es visible para su dueño y el admin; `removed` es visible para el admin y para los participantes históricos de una conversación.

### 3. Precio numérico en pesos colombianos (COP)

Se agrega `price_cents BIGINT` como unidad mínima del peso colombiano y se conserva temporalmente `price TEXT` como respaldo de migración.

- **Por qué:** permite filtrar, ordenar y validar sin conversiones frágiles en SQL.
- **Moneda:** todos los montos de la plataforma son pesos colombianos (COP). El formulario y los filtros se etiquetan explícitamente “(COP)”.
- **Backfill:** los valores existentes se interpretan como pesos colombianos enteros, eliminando separadores de miles y multiplicando por 100. Los valores no parseables se detectan antes de imponer `NOT NULL`.
- **Entrada:** el formulario recibe pesos enteros, admite separadores de miles y rechaza valores no numéricos, cero, negativos o con decimales.
- **Formato:** la vista muestra `$ 45.000.000` usando el separador de miles colombiano (`.`) y sin decimales. El sufijo `COP` se usa en etiquetas de formulario y filtros para evitar ambigüedad.

### 4. Búsqueda con `ILIKE` parametrizado e índices

El backend divide `q` en tokens (máximo 5), y por cada token agrega una condición parametrizada sobre marca, modelo, descripción y año. Los filtros de año y rango de precio se combinan con `AND`. La consulta usa `LIMIT/OFFSET` con página de 12 y una consulta de conteo con el mismo `WHERE`.

- **Por qué:** satisface `Mazda 2013` y búsquedas por descripción sin depender de un diccionario en español ni de extensiones.
- **Alternativas descartadas:** `tsvector` con GIN es más escalable pero añade configuración de idioma y complejidad de mantenimiento; `pg_trgm` mejora `ILIKE` pero requiere una extensión que puede no estar disponible en todos los proveedores.
- **Rendimiento:** índices btree en `(status, created_at)`, `(status, year)`, `(status, price_cents)` y `(seller_id)`. Con el volumen esperado, el filtro `ILIKE` sobre un conjunto activo pequeño es aceptable; si crece, se añadirá `pg_trgm` como evolución.
- **Seguridad:** los valores siempre viajan como parámetros; la estructura SQL se construye con un número fijo de fragmentos, nunca concatenando texto del usuario.

### 5. Conversación única por publicación e interesado

`conversations` tiene una restricción única `(car_id, buyer_id)` y `CHECK (buyer_id <> seller_id)`. Al contactar de nuevo, se reutiliza el hilo y se agrega un mensaje.

- **Por qué:** evita hilos duplicados y simplifica la bandeja.
- **Permisos:** la consulta de conversación filtra por `buyer_id` o `seller_id`; el administrador usa una consulta de lectura separada. La autorización se valida en el servicio, no solo en la ruta.
- **Integridad:** `car_id` usa `ON DELETE RESTRICT` y las publicaciones nunca se borran físicamente; así el hilo conserva su referencia.

### 6. Mensajería solo para cuentas existentes

El botón de interés exige sesión. Si el visitante no la tiene, se le redirige a login/registro con un parámetro `next` validado como ruta local. El texto editado se conserva en `sessionStorage` por publicación y se restaura al volver.

- **Por qué:** el modelo actual solo registra vendedores y no hay identidad anónima segura; crear un rol de comprador queda fuera de alcance.
- **Consecuencia:** quien quiera contactar debe registrarse como vendedor. Se documenta como limitación de producto y se deja la puerta abierta a un rol de comprador futuro.
- **Alternativa descartada:** mensajes anónimos; exigirían identidad temporal, moderación de spam y reglas de privacidad adicionales.

### 7. Ventana de edición de diez minutos en el servidor

El servicio compara `created_at` con `clock.Now().Add(-10 * time.Minute)` y pasa el corte al repositorio. La actualización incluye `status='active'`, `seller_id` del usuario y `created_at >= cutoff`; si no afecta filas, se distingue entre “no existe/no es tuyo” y “plazo expirado”.

- **Por qué:** el reloj del navegador y la ocultación del botón no son fuentes de verdad.
- **UI:** el enlace de edición se muestra solo si la publicación está activa y dentro del plazo; si el usuario llega por URL directa, recibe un mensaje claro.
- **Retirada:** no tiene límite temporal; solo la edición está limitada.

### 8. Carga de fotos y servicio con caché

- **Límites:** 5 MB por archivo, 6 fotos por publicación, JPEG/PNG, dimensiones máximas 6000×6000 y mínimas razonables para evitar imágenes vacías.
- **Validación:** `http.MaxBytesReader` sobre el cuerpo, `http.DetectContentType` sobre los primeros 512 bytes y `image.DecodeConfig` para confirmar que el contenido es una imagen decodificable. No se confía en la extensión ni en el `Content-Type` del cliente.
- **Almacenamiento:** se genera un ID propio, se calcula SHA-256 y se guarda el binario normalizado; el nombre original solo se conserva como metadato saneado, nunca como clave.
- **Servicio:** `GET /photos/:id` carga solo la fila solicitada, verifica el estado de la publicación y responde con `Content-Type`, `Content-Length`, `ETag` (SHA-256) y `Cache-Control`. Se usa `http.ServeContent` para soportar solicitudes condicionales y rangos. Las fotos de publicaciones no activas solo son accesibles para el dueño o un administrador.
- **Catálogo:** las tarjetas usan `loading="lazy"` y la URL del endpoint; nunca se incrustan binarios en el HTML.

### 9. Idempotencia de publicación

El formulario incluye un `client_token` aleatorio generado al renderizarse y una restricción única parcial en `cars`. Si el mismo formulario se envía dos veces, la segunda inserción no crea otra fila y redirige a la publicación existente del mismo vendedor.

- **Por qué:** resuelve el doble clic sin depender solo de JavaScript.
- **Alternativa descartada:** deshabilitar el botón en el cliente; útil como refuerzo, pero insuficiente si la petición se repite.

### 10. Autorización en capas

- Rutas: grupos con `RequireRole` para vendedor y administrador.
- Servicio/repositorio: verificación de propiedad y de participación en la consulta.
- Respuestas: `401` sin sesión, `403` sin permiso o propiedad, `409` para plazo expirado o conflicto de estado, `413` para archivos demasiado grandes, `422` para validación.
- CSRF: se añade un token de doble envío (cookie no HttpOnly + campo oculto) y se valida en todos los métodos POST. El logout pasa a POST; se conserva una ruta GET solo como confirmación/redirección si fuera necesario para compatibilidad.

### 11. Templates con partials y nueva identidad visual

Se agrega `partials.html` con fragmentos de encabezado, navegación, pie, mensajes flash y campo CSRF. Las páginas lo incluyen con `{{template "..." .}}`; `template.ParseFS` ya carga todos los archivos embebidos.

La paleta se define en variables CSS:

```css
--bg: #f5f2ee;        /* blanco cálido */
--surface: #ffffff;
--ink: #1e1c1a;       /* carbón */
--muted: #6b655f;
--border: #e3ddd5;
--accent: #a8522a;    /* terracota */
--accent-strong: #8a3f1f;
--danger: #a12b22;
--success: #2f6b4f;
```

No se usa azul como color principal ni de acento. El encabezado y el pie usan carbón; las tarjetas priorizan la fotografía con relación de aspecto fija y el precio con tipografía destacada. La cuadrícula pasa de tres columnas a una en móvil y se añaden estados de foco visibles.

### 12. Migraciones idempotentes y limpieza de datos demo

Se mantiene el arreglo de statements idempotentes aplicado dentro de una transacción con advisory lock. Las nuevas sentencias son aditivas:

- `ALTER TABLE cars ADD COLUMN IF NOT EXISTS status ...`
- `ALTER TABLE cars ADD COLUMN IF NOT EXISTS price_cents BIGINT`
- `ALTER TABLE cars ADD COLUMN IF NOT EXISTS client_token TEXT`
- `ALTER TABLE cars ADD COLUMN IF NOT EXISTS edited_at TIMESTAMPTZ`
- Creación de `car_photos`, `conversations`, `messages`, `moderation_logs` e índices.
- Backfill de `price_cents`.
- Eliminación idempotente de los seis vehículos demo conocidos (`seller_id IS NULL` y combinaciones exactas de marca/modelo/año).
- `SET NOT NULL` de `seller_id` y `price_cents` solo después de la limpieza y el backfill.

Se conserva `price TEXT` hasta verificar la migración; una tarea posterior puede retirarla. El rollback inmediato consiste en desplegar la versión anterior: las columnas nuevas son ignoradas y `price` sigue existiendo. Si el `SET NOT NULL` falla por datos heredados no parseables, la transacción se revierte y el despliegue se detiene con un error explícito, sin dejar el esquema a medias.

### 13. Eliminación del seeding

Se elimina la función `seedCars` y su llamada en `internal/app`. La limpieza de los registros ya sembrados se hace en la migración, no en el arranque, para que sea idempotente y quede registrada.

## Risks / Trade-offs

- **Crecimiento de PostgreSQL por imágenes** → límites de 5 MB y 6 fotos, normalización a JPEG de 1600 px, caché inmutable y endpoint separado.
- **Consumo de memoria al servir fotos en serverless** → se carga solo una fila por solicitud, se usan límites de tamaño y se evita seleccionar blobs en listados.
- **`ILIKE` no escala indefinidamente** → índices por estado y columnas filtradas; migración futura a `pg_trgm` o búsqueda de texto completo si el volumen lo exige.
- **Diferencia de reloj entre aplicación y base de datos en la ventana de 10 minutos** → el corte se calcula con el `Clock` de la aplicación y se compara contra `created_at` en UTC; se documenta que la ventana es aproximada al segundo.
- **Pérdida de datos heredados al imponer `NOT NULL`** → backfill verificable, conservación temporal de `price`, transacción única y mensaje de error explícito.
- **Valores heredados expresados en otra moneda** → los registros demo se eliminan; cualquier precio real no parseable como COP se detecta antes de imponer `NOT NULL` y se revisa manualmente antes de desplegar.
- **Limpieza de demo podría afectar datos reales** → se limita a filas con `seller_id IS NULL` y combinaciones exactas de los seis vehículos sembrados.
- **CSRF y logout por GET** → token de doble envío, POST para cambios de estado y SameSite Lax ya existente.
- **Fotos sin miniaturas dedicadas** → la normalización acota el peso; si el catálogo crece se añadirá una miniatura específica.
- **Registrar compradores como vendedores para contactar** → limitación aceptada; se documenta y se deja preparado el modelo de roles para un comprador futuro.

## Migration Plan

1. Ejecutar los tests actuales y respaldar la base de datos.
2. Desplegar la migración aditiva: columnas, tablas e índices.
3. Backfill de `price_cents` en COP y verificación de valores nulos o no parseables.
4. Eliminar los seis vehículos demo conocidos.
5. Imponer `NOT NULL` en `seller_id` y `price_cents`.
6. Desplegar el código que usa `price_cents`, estados y nuevas rutas.
7. Verificar catálogo vacío, publicación, búsqueda, mensajería y moderación en un entorno de staging.
8. Rollback: volver al despliegue anterior; `price` y las columnas antiguas siguen presentes. Si hiciera falta revertir DDL, eliminar primero los índices y tablas nuevas y después las columnas agregadas, con respaldo previo.
