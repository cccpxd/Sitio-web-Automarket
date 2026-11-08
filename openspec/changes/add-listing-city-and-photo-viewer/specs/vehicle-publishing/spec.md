# Spec Delta

## MODIFIED Requirements

### Requirement: Formulario de publicación

El sistema SHALL ofrecer a los vendedores autenticados un formulario con marca, modelo, año, precio en pesos colombianos (COP), ciudad, descripción y una o varias fotografías.

#### Scenario: Vendedor abre el formulario

- **WHEN** un vendedor autenticado accede al apartado de publicación
- **THEN** el sistema muestra los campos de marca, modelo, año, precio etiquetado como “(COP)”, ciudad, descripción y carga de fotografías

#### Scenario: Visitante o administrador intenta publicar

- **WHEN** un visitante sin sesión o un administrador intenta crear una publicación
- **THEN** el sistema rechaza la operación

### Requirement: Validación de campos obligatorios

El sistema SHALL validar en el servidor que marca, modelo, año, precio, ciudad y al menos una fotografía estén presentes antes de guardar la publicación.

#### Scenario: Campos incompletos

- **WHEN** el vendedor envía el formulario sin marca, modelo, año, precio, ciudad o fotografías
- **THEN** el sistema rechaza la publicación y muestra mensajes de validación claros

#### Scenario: Datos válidos

- **WHEN** el vendedor envía todos los campos obligatorios, incluida la ciudad, con formatos válidos
- **THEN** el sistema acepta la publicación

### Requirement: Edición dentro del plazo

El sistema SHALL permitir al vendedor propietario editar los datos, incluida la ciudad, y las fotografías de su publicación solo dentro de los primeros diez minutos desde su creación, y SHALL rechazar la edición fuera del plazo.

#### Scenario: Edición válida

- **WHEN** el propietario edita su publicación dentro de los diez minutos y envía datos válidos, incluida la ciudad
- **THEN** el sistema actualiza los datos, la ciudad y las fotografías y registra la fecha de edición

#### Scenario: Edición fuera del plazo

- **WHEN** el propietario intenta editar después de diez minutos
- **THEN** el sistema rechaza los cambios e informa que el plazo expiró

#### Scenario: Edición de publicación retirada

- **WHEN** el propietario intenta editar una publicación retirada
- **THEN** el sistema rechaza la edición

## ADDED Requirements

### Requirement: Validación de ciudad

El sistema SHALL aceptar como ciudad un texto libre no vacío, normalizado en espacios y con un máximo de 80 caracteres, y SHALL rechazar una ciudad vacía o demasiado larga.

#### Scenario: Ciudad válida

- **WHEN** el vendedor envía una ciudad con espacios sobrantes, como `  Bogotá  `
- **THEN** el sistema la guarda normalizada como `Bogotá`

#### Scenario: Ciudad vacía

- **WHEN** el vendedor envía el formulario sin ciudad o con solo espacios
- **THEN** el sistema rechaza la publicación y muestra el error junto al campo de ciudad

#### Scenario: Ciudad demasiado larga

- **WHEN** el vendedor envía una ciudad de más de 80 caracteres
- **THEN** el sistema rechaza la publicación e indica el límite

### Requirement: Ciudad en publicaciones anteriores al cambio

El sistema SHALL mostrar las publicaciones creadas antes de la incorporación de la ciudad como “Ciudad no especificada” y SHALL NOT inventar ni rellenar una ubicación para ellas.

#### Scenario: Publicación heredada sin ciudad

- **WHEN** un visitante consulta una publicación creada antes del cambio
- **THEN** el sistema la muestra como “Ciudad no especificada” sin asignarle una ciudad ficticia

#### Scenario: Edición dentro del plazo de una publicación heredada

- **WHEN** el propietario edita dentro de los diez minutos una publicación sin ciudad
- **THEN** el formulario exige una ciudad válida antes de guardar

#### Scenario: Publicación heredada fuera del plazo

- **WHEN** una publicación sin ciudad supera el plazo de edición
- **THEN** conserva su registro sin ciudad y el sistema no ofrece una acción especial para completarla
