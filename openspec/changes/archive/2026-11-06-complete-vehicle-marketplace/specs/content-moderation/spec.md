# Spec Delta

## Purpose

Permite al administrador retirar publicaciones que incumplen las reglas después de su publicación, exigiendo una razón y conservando evidencia auditable de la acción.

## ADDED Requirements

### Requirement: Moderación posterior a la publicación

El sistema SHALL publicar los vehículos de inmediato y SHALL NOT exigir aprobación previa del administrador.

#### Scenario: Publicación sin aprobación

- **WHEN** un vendedor guarda una publicación válida
- **THEN** el sistema la muestra públicamente sin esperar una revisión administrativa

### Requirement: Retirada administrativa con razón obligatoria

El administrador SHALL poder retirar cualquier publicación y el sistema SHALL exigir una razón no vacía antes de aplicar la retirada.

#### Scenario: Retirada con razón

- **WHEN** el administrador envía una retirada con una razón válida
- **THEN** el sistema retira la publicación y registra la razón

#### Scenario: Retirada sin razón

- **WHEN** el administrador intenta retirar una publicación sin indicar razón
- **THEN** el sistema rechaza la operación y no modifica la publicación

### Requirement: Confirmación antes de retirar

El sistema SHALL solicitar una confirmación explícita antes de ejecutar la retirada de una publicación.

#### Scenario: Confirmación de retirada

- **WHEN** el administrador selecciona retirar una publicación
- **THEN** el sistema muestra una pantalla de confirmación con el campo de razón

#### Scenario: Confirmación cancelada

- **WHEN** el administrador cancela la confirmación
- **THEN** la publicación permanece en su estado anterior

### Requirement: Registro de auditoría

El sistema SHALL registrar de forma inmutable el administrador responsable, la publicación afectada, la razón y la fecha de cada retirada administrativa.

#### Scenario: Auditoría creada

- **WHEN** se completa una retirada administrativa
- **THEN** existe un registro con administrador, publicación, razón y fecha

#### Scenario: Evidencia conservada

- **WHEN** el administrador consulta el registro de moderación
- **THEN** el sistema muestra las retiradas históricas con sus datos aunque la publicación ya no sea pública

#### Scenario: Intento de eliminar la evidencia

- **WHEN** se retira o consulta una publicación
- **THEN** el sistema no elimina el registro de auditoría asociado

### Requirement: Ocultamiento de la publicación retirada

Una publicación retirada por el administrador SHALL dejar de aparecer en el catálogo público, en Ventas, en el detalle y en los resultados del buscador.

#### Scenario: Catálogo después de la retirada

- **WHEN** un administrador retira una publicación
- **THEN** la publicación deja de mostrarse en el catálogo, Ventas y la búsqueda

#### Scenario: Acceso directo a la publicación retirada

- **WHEN** un visitante accede directamente a la URL de una publicación retirada
- **THEN** el sistema no muestra la publicación como activa

### Requirement: Borrado lógico e integridad

El sistema SHALL conservar el registro de la publicación mediante un estado de retirada y SHALL NOT eliminar físicamente la fila, preservando mensajes y referencias.

#### Scenario: Publicación con conversaciones

- **WHEN** el administrador retira una publicación que tiene mensajes asociados
- **THEN** la publicación y sus conversaciones siguen existiendo y los mensajes no se pierden

### Requirement: Acceso exclusivo del administrador

Solo un administrador autenticado SHALL poder retirar publicaciones y consultar el registro de moderación.

#### Scenario: Vendedor intenta retirar una publicación ajena

- **WHEN** un vendedor intenta ejecutar la retirada administrativa de una publicación
- **THEN** el sistema rechaza la operación

#### Scenario: Visitante consulta el registro

- **WHEN** un visitante sin sesión intenta abrir el registro de moderación
- **THEN** el sistema le exige autenticación y no muestra la información
