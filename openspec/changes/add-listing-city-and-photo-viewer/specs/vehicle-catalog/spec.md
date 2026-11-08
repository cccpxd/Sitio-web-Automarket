# Spec Delta

## MODIFIED Requirements

### Requirement: Contenido de cada publicación en el catálogo

Cada tarjeta del catálogo SHALL mostrar al menos una fotografía, marca, modelo, año, precio en pesos colombianos (COP), ciudad y una descripción breve, priorizando la fotografía y el precio.

#### Scenario: Tarjeta de vehículo

- **WHEN** el catálogo renderiza una publicación activa
- **THEN** la tarjeta muestra su fotografía principal, marca, modelo, año, ciudad y el precio formateado en COP con separador de miles y sin decimales (por ejemplo, `$ 45.000.000`)

#### Scenario: Publicación sin ciudad registrada

- **WHEN** una publicación anterior a la incorporación de la ciudad aparece en el catálogo
- **THEN** la tarjeta muestra “Ciudad no especificada” en lugar de inventar una ubicación

#### Scenario: Publicación sin fotografía

- **WHEN** una publicación no tiene fotografías asociadas
- **THEN** el sistema no la muestra como publicación válida en el catálogo

### Requirement: Detalle del vehículo

El sistema SHALL ofrecer una página de detalle para cada publicación activa con todas sus fotografías, marca, modelo, año, precio en COP, ciudad, descripción completa, fecha de publicación e información pública del vendedor.

#### Scenario: Visitante abre el detalle

- **WHEN** un visitante accede a la URL de detalle de una publicación activa
- **THEN** el sistema muestra todas sus fotografías y datos, incluidos el nombre público del vendedor, la fecha de publicación y la ciudad

#### Scenario: Detalle de publicación sin ciudad

- **WHEN** un visitante accede al detalle de una publicación activa creada antes de la incorporación de la ciudad
- **THEN** el sistema muestra “Ciudad no especificada” y no inventa una ubicación

#### Scenario: Detalle de publicación inexistente o no activa

- **WHEN** un visitante accede al detalle de una publicación inexistente, retirada o no activa
- **THEN** el sistema responde con un estado de no encontrado o una página informativa, sin revelar datos privados

## ADDED Requirements

### Requirement: Visor de fotografías en la misma página

El sistema SHALL mostrar las fotografías del detalle en un visor dentro de la misma página y SHALL NOT abrir una pestaña o ventana nueva al activarlas.

#### Scenario: Abrir el visor desde la fotografía principal

- **WHEN** un visitante activa la fotografía principal del detalle
- **THEN** el sistema abre un visor modal con esa fotografía sin abrir una pestaña o ventana nueva

#### Scenario: Abrir el visor desde una miniatura

- **WHEN** un visitante activa una miniatura de la galería
- **THEN** el sistema abre el visor en la posición de esa fotografía

#### Scenario: Navegar entre fotografías

- **WHEN** el visor está abierto y el visitante usa los controles anterior o siguiente, o las flechas del teclado
- **THEN** el sistema muestra la fotografía correspondiente y actualiza el contador de posición

#### Scenario: Cerrar el visor

- **WHEN** el visitante usa el botón de cerrar, presiona `Esc` o activa el fondo del visor
- **THEN** el visor se cierra y el detalle vuelve a ser la vista activa

#### Scenario: Publicación con una sola fotografía

- **WHEN** el visitante abre el visor de una publicación con una sola fotografía
- **THEN** el visor muestra la imagen sin controles de navegación anterior o siguiente

#### Scenario: Visor sin JavaScript

- **WHEN** el navegador no ejecuta JavaScript y el visitante activa una fotografía
- **THEN** la fotografía se abre en la misma pestaña, sin abrir una ventana nueva

### Requirement: Accesibilidad del visor de fotografías

El sistema SHALL permitir operar el visor con teclado, mantener el foco dentro del diálogo mientras esté abierto, anunciar la posición de la fotografía y devolver el foco al control que lo abrió al cerrar.

#### Scenario: Navegación por teclado

- **WHEN** el visor está abierto y el visitante usa `Tab`, `Shift+Tab`, las flechas o `Esc`
- **THEN** el foco recorre los controles del visor, cambia de fotografía o cierra el visor según la tecla, sin escapar al contenido de fondo

#### Scenario: Retorno del foco

- **WHEN** el visitante cierra el visor
- **THEN** el foco vuelve a la fotografía principal o a la miniatura que abrió el visor

#### Scenario: Anuncio de la posición

- **WHEN** el visitante cambia de fotografía dentro del visor
- **THEN** el sistema anuncia la posición actual y el total de fotografías a las tecnologías de asistencia
