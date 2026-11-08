# Spec Delta

## Purpose

Permite a los vendedores crear y editar publicaciones completas de vehículos con datos validados y fotografías, sin aprobación previa.

## ADDED Requirements

### Requirement: Formulario de publicación

El sistema SHALL ofrecer a los vendedores autenticados un formulario con marca, modelo, año, precio en pesos colombianos (COP), descripción y una o varias fotografías.

#### Scenario: Vendedor abre el formulario

- **WHEN** un vendedor autenticado accede al apartado de publicación
- **THEN** el sistema muestra los campos de marca, modelo, año, precio etiquetado como “(COP)”, descripción y carga de fotografías

#### Scenario: Visitante o administrador intenta publicar

- **WHEN** un visitante sin sesión o un administrador intenta crear una publicación
- **THEN** el sistema rechaza la operación

### Requirement: Validación de campos obligatorios

El sistema SHALL validar en el servidor que marca, modelo, año, precio y al menos una fotografía estén presentes antes de guardar la publicación.

#### Scenario: Campos incompletos

- **WHEN** el vendedor envía el formulario sin marca, modelo, año, precio o fotografías
- **THEN** el sistema rechaza la publicación y muestra mensajes de validación claros

#### Scenario: Datos válidos

- **WHEN** el vendedor envía todos los campos obligatorios con formatos válidos
- **THEN** el sistema acepta la publicación

### Requirement: Validación de año y precio

El sistema SHALL aceptar únicamente años dentro del rango permitido y precios enteros en COP mayores que cero, rechazando valores vacíos, negativos, no numéricos o con decimales.

#### Scenario: Año inválido

- **WHEN** el vendedor envía un año anterior a 1900 o posterior al año máximo permitido
- **THEN** el sistema rechaza la publicación e indica el rango válido

#### Scenario: Precio inválido

- **WHEN** el vendedor envía un precio vacío, negativo, cero, no numérico o con decimales
- **THEN** el sistema rechaza la publicación e indica que el precio debe ser un valor entero en COP

### Requirement: Fotografías con vista previa y eliminación

El sistema SHALL permitir seleccionar una o varias fotografías, mostrar una vista previa antes de publicar y quitar o cambiar las fotografías seleccionadas antes de guardar.

#### Scenario: Vista previa de fotografías

- **WHEN** el vendedor selecciona fotografías válidas en el formulario
- **THEN** el sistema muestra una vista previa de cada una antes de enviar el formulario

#### Scenario: Quitar fotografía antes de guardar

- **WHEN** el vendedor quita una fotografía de la selección previa
- **THEN** esa fotografía no se guarda con la publicación

### Requirement: Publicación atómica y sin duplicados

El sistema SHALL guardar la publicación y sus fotografías como una sola operación y SHALL evitar publicaciones duplicadas causadas por varios clics o envíos repetidos del mismo formulario.

#### Scenario: Doble clic en publicar

- **WHEN** el vendedor envía dos veces el mismo formulario por un doble clic
- **THEN** el sistema crea una sola publicación y redirige al resultado existente

#### Scenario: Fallo al guardar una fotografía

- **WHEN** una de las fotografías no puede guardarse
- **THEN** el sistema no deja una publicación parcial y muestra un error

### Requirement: Edición dentro del plazo

El sistema SHALL permitir al vendedor propietario editar los datos y las fotografías de su publicación solo dentro de los primeros diez minutos desde su creación, y SHALL rechazar la edición fuera del plazo.

#### Scenario: Edición válida

- **WHEN** el propietario edita su publicación dentro de los diez minutos y envía datos válidos
- **THEN** el sistema actualiza los datos y las fotografías y registra la fecha de edición

#### Scenario: Edición fuera del plazo

- **WHEN** el propietario intenta editar después de diez minutos
- **THEN** el sistema rechaza los cambios e informa que el plazo expiró

#### Scenario: Edición de publicación retirada

- **WHEN** el propietario intenta editar una publicación retirada
- **THEN** el sistema rechaza la edición

### Requirement: Conservación de datos ante errores

Cuando una publicación es rechazada, el sistema SHALL conservar los datos ya introducidos por el vendedor y mostrar los errores junto a los campos correspondientes.

#### Scenario: Error de validación en el formulario

- **WHEN** el sistema rechaza el formulario por un dato inválido
- **THEN** la respuesta mantiene los valores válidos introducidos y resalta el campo con error
