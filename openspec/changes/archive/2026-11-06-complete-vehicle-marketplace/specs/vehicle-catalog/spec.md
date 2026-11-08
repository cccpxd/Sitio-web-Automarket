# Spec Delta

## Purpose

Proporciona un catálogo público de vehículos activos con información completa, detalle por publicación y estados vacíos claros, sin datos ficticios.

## ADDED Requirements

### Requirement: Catálogo público de publicaciones activas

El sistema SHALL mostrar en el catálogo únicamente publicaciones con estado activo y SHALL permitir consultarlo sin iniciar sesión.

#### Scenario: Catálogo con publicaciones activas

- **WHEN** un visitante abre el catálogo
- **THEN** el sistema lista las publicaciones activas y excluye las retiradas por el vendedor o por el administrador

#### Scenario: Publicación retirada

- **WHEN** una publicación cambia a un estado no activo
- **THEN** deja de aparecer en el catálogo y en los resultados del buscador

### Requirement: Contenido de cada publicación en el catálogo

Cada tarjeta del catálogo SHALL mostrar al menos una fotografía, marca, modelo, año, precio en pesos colombianos (COP) y una descripción breve, priorizando la fotografía y el precio.

#### Scenario: Tarjeta de vehículo

- **WHEN** el catálogo renderiza una publicación activa
- **THEN** la tarjeta muestra su fotografía principal, marca, modelo, año y el precio formateado en COP con separador de miles y sin decimales (por ejemplo, `$ 45.000.000`)

#### Scenario: Publicación sin fotografía

- **WHEN** una publicación no tiene fotografías asociadas
- **THEN** el sistema no la muestra como publicación válida en el catálogo

### Requirement: Detalle del vehículo

El sistema SHALL ofrecer una página de detalle para cada publicación activa con todas sus fotografías, marca, modelo, año, precio en COP, descripción completa, fecha de publicación e información pública del vendedor.

#### Scenario: Visitante abre el detalle

- **WHEN** un visitante accede a la URL de detalle de una publicación activa
- **THEN** el sistema muestra todas sus fotografías y datos, incluido el nombre público del vendedor y la fecha de publicación

#### Scenario: Detalle de publicación inexistente o no activa

- **WHEN** un visitante accede al detalle de una publicación inexistente, retirada o no activa
- **THEN** el sistema responde con un estado de no encontrado o una página informativa, sin revelar datos privados

### Requirement: Visibilidad inmediata

Una publicación SHALL aparecer en el catálogo público y en el apartado Ventas inmediatamente después de guardarse correctamente, sin aprobación previa del administrador.

#### Scenario: Publicación recién creada

- **WHEN** un vendedor guarda una publicación válida
- **THEN** la publicación aparece de inmediato para visitantes y vendedores

### Requirement: Estado vacío sin datos ficticios

El sistema SHALL iniciar con el catálogo vacío cuando no existan publicaciones reales, mostrar un mensaje informativo como “Todavía no hay vehículos publicados” y no SHALL sembrar vehículos, vendedores ni mensajes de ejemplo.

#### Scenario: Catálogo sin publicaciones

- **WHEN** no existe ninguna publicación activa
- **THEN** el catálogo muestra un estado vacío informativo y no muestra vehículos de ejemplo

#### Scenario: Base de datos nueva

- **WHEN** la aplicación arranca contra una base de datos sin publicaciones
- **THEN** el sistema no inserta datos ficticios automáticamente

### Requirement: Fotografías públicas asociadas

El sistema SHALL mostrar las fotografías únicamente mientras la publicación esté activa y SHALL impedir el acceso público a las fotografías de publicaciones retiradas.

#### Scenario: Fotografía de publicación activa

- **WHEN** se solicita una fotografía de una publicación activa
- **THEN** el sistema la entrega con su tipo de contenido correcto

#### Scenario: Fotografía de publicación retirada

- **WHEN** un visitante solicita una fotografía de una publicación retirada
- **THEN** el sistema deniega el acceso público
