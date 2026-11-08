# visual-identity Specification

## Purpose
Define una identidad visual formal, limpia y consistente, sin azul como color principal ni de acento, adaptada a móviles, tabletas y escritorio.

## Requirements

### Requirement: Paleta sin azul dominante

La interfaz SHALL usar una paleta sobria basada en tonos carbón, blanco y gris con un acento cálido discreto, y SHALL NOT usar azul como color principal ni como color de acento.

#### Scenario: Revisión de la paleta

- **WHEN** se inspeccionan los colores principales y de acento de la interfaz
- **THEN** ninguno de ellos es azul y la paleta corresponde a carbón, blanco, gris y un acento cálido

#### Scenario: Componentes principales

- **WHEN** se renderizan encabezado, botones, enlaces y tarjetas
- **THEN** todos usan la paleta definida y no introducen azul como color dominante

### Requirement: Contraste y legibilidad

La interfaz SHALL mantener un contraste suficiente entre texto y fondo y SHALL usar tamaños y espaciados que faciliten la lectura.

#### Scenario: Texto sobre fondo

- **WHEN** se muestra texto principal, secundario o de validación
- **THEN** su contraste cumple un nivel de accesibilidad AA

#### Scenario: Navegación por teclado

- **WHEN** el usuario navega con teclado por enlaces, campos y botones
- **THEN** el elemento enfocado se distingue visualmente

### Requirement: Jerarquía visual

La interfaz SHALL establecer una jerarquía clara para títulos, precios, datos técnicos, etiquetas y botones.

#### Scenario: Detalle de vehículo

- **WHEN** se muestra el detalle de una publicación
- **THEN** el precio y el título del vehículo se identifican como los elementos de mayor jerarquía después de la fotografía

#### Scenario: Formulario de publicación

- **WHEN** se muestra el formulario de publicación
- **THEN** cada campo tiene una etiqueta visible y los errores se asocian al campo correspondiente

### Requirement: Tarjetas centradas en fotografía y precio

Las tarjetas de vehículos SHALL dar prioridad visual a la fotografía y al precio sobre el resto de los datos.

#### Scenario: Tarjeta en el catálogo

- **WHEN** se renderiza una tarjeta de vehículo
- **THEN** la fotografía ocupa un área principal y el precio se muestra con énfasis

### Requirement: Diseño responsive

La interfaz SHALL adaptarse correctamente a móviles, tabletas y escritorio, sin pérdida de contenido ni controles inaccesibles.

#### Scenario: Vista móvil

- **WHEN** se abre el catálogo, el detalle, el formulario, Ventas, Mensajes o administración en un ancho de móvil
- **THEN** el contenido se reorganiza en una sola columna y todas las acciones siguen siendo accesibles

#### Scenario: Vista de escritorio

- **WHEN** se abre la misma pantalla en escritorio
- **THEN** el contenido aprovecha el ancho disponible con una cuadrícula legible

### Requirement: Consistencia entre pantallas

El sistema SHALL mantener la misma identidad, componentes y patrones visuales en catálogo, detalle, publicación, Ventas, Mensajes y administración.

#### Scenario: Navegación entre secciones

- **WHEN** el usuario pasa del catálogo al detalle, al formulario, a Ventas o a Mensajes
- **THEN** encabezado, navegación, tipografía, botones y colores se mantienen consistentes

### Requirement: Mensajes de validación claros

La interfaz SHALL mostrar etiquetas y mensajes de validación comprensibles, indicando qué campo falló y cómo corregirlo.

#### Scenario: Error en formulario

- **WHEN** el usuario envía un formulario con datos inválidos
- **THEN** el sistema muestra un mensaje claro junto al campo afectado y conserva el resto de los datos

### Requirement: Estados vacíos y de error

La interfaz SHALL incluir estados visuales claros para listados vacíos, errores y confirmaciones, sin depender de efectos innecesarios.

#### Scenario: Catálogo vacío

- **WHEN** no hay publicaciones activas
- **THEN** la interfaz muestra un estado vacío informativo y coherente con el diseño general

#### Scenario: Confirmación de retirada

- **WHEN** se solicita retirar una publicación
- **THEN** la confirmación se presenta con un estilo consistente y sin animaciones innecesarias
