# Spec Delta

## Purpose

Ofrece al vendedor un apartado Ventas basado en las mismas publicaciones activas del catálogo público, junto con la gestión de sus propias publicaciones.

## ADDED Requirements

### Requirement: Ventas usa la misma fuente que el catálogo

El apartado Ventas SHALL mostrar las publicaciones activas usando la misma fuente de datos y las mismas reglas de visibilidad que el catálogo público, evitando listados distintos o desactualizados.

#### Scenario: Vendedor abre Ventas

- **WHEN** un vendedor autenticado accede a Ventas
- **THEN** el sistema le muestra las mismas publicaciones activas que ve un visitante en el catálogo

#### Scenario: Publicación nueva visible en Ventas

- **WHEN** cualquier vendedor publica un vehículo válido
- **THEN** la publicación aparece en Ventas para los demás vendedores y en el catálogo público

#### Scenario: Publicación retirada

- **WHEN** una publicación es retirada por su vendedor o por el administrador
- **THEN** deja de aparecer en Ventas para el resto de usuarios

### Requirement: Consistencia entre catálogo y Ventas

El sistema SHALL mantener el mismo conjunto de vehículos activos en el catálogo público y en Ventas.

#### Scenario: Comparación de listados

- **WHEN** se comparan las publicaciones activas del catálogo y las de Ventas
- **THEN** ambos listados contienen exactamente las mismas publicaciones activas

### Requirement: Bloque de publicaciones propias

El apartado Ventas SHALL identificar y listar por separado las publicaciones del vendedor autenticado, mostrando su estado y las acciones disponibles.

#### Scenario: Vendedor consulta sus publicaciones

- **WHEN** el vendedor abre Ventas
- **THEN** ve sus publicaciones activas y retiradas con su estado correspondiente

#### Scenario: Vendedor no ve retiradas ajenas

- **WHEN** otro vendedor retira una publicación
- **THEN** el vendedor autenticado no la ve en su bloque de publicaciones propias

### Requirement: Edición temporal desde Ventas

El sistema SHALL mostrar la acción de editar únicamente mientras la publicación esté activa y dentro de los primeros diez minutos desde su creación, y SHALL rechazar en el servidor cualquier intento fuera de esas condiciones.

#### Scenario: Acción de editar visible

- **WHEN** el vendedor consulta una publicación propia activa creada hace menos de diez minutos
- **THEN** el sistema muestra la acción de editar

#### Scenario: Acción de editar oculta o deshabilitada

- **WHEN** el vendedor consulta una publicación propia creada hace más de diez minutos o en estado retirado
- **THEN** el sistema no le ofrece editar y, si intenta el endpoint directamente, lo rechaza

### Requirement: Retirada por el vendedor

El sistema SHALL permitir al vendedor retirar su propia publicación en cualquier momento, previa confirmación, sin exigir razón y sin generar un registro de moderación administrativa.

#### Scenario: Retirada confirmada

- **WHEN** el vendedor confirma la retirada de su publicación
- **THEN** la publicación cambia a estado retirado y deja de aparecer en el catálogo, la búsqueda y Ventas pública

#### Scenario: Retirada cancelada

- **WHEN** el vendedor abre la confirmación y cancela
- **THEN** la publicación permanece activa

#### Scenario: Conversaciones asociadas

- **WHEN** el vendedor retira una publicación que tiene conversaciones
- **THEN** las conversaciones y sus mensajes se conservan y siguen siendo accesibles para sus participantes

### Requirement: Publicaciones retiradas del vendedor

El sistema SHALL conservar las publicaciones retiradas por el vendedor en su bloque de publicaciones propias, con un estado visible que indique que no son públicas.

#### Scenario: Publicación retirada en el bloque propio

- **WHEN** el vendedor consulta una publicación que retiró
- **THEN** la ve con el estado “Retirada” y sin enlace público activo
