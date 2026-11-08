# Spec Delta

## Purpose

Garantiza que las fotografías de las publicaciones se almacenen, validen y sirvan de forma segura y eficiente desde PostgreSQL.

## ADDED Requirements

### Requirement: Almacenamiento en PostgreSQL

El sistema SHALL almacenar el contenido binario de cada fotografía en PostgreSQL, en una tabla separada de la publicación, y SHALL NOT guardar imágenes codificadas en Base64 dentro de campos de texto.

#### Scenario: Fotografía guardada

- **WHEN** el vendedor adjunta una fotografía válida a una publicación
- **THEN** el sistema persiste el contenido binario asociado a la publicación y a su posición

#### Scenario: Imagen en campo de texto

- **WHEN** se guarda una publicación
- **THEN** el sistema no almacena la imagen como texto Base64

### Requirement: Validación del tipo real de archivo

El sistema SHALL determinar el tipo de archivo por su contenido real y SHALL aceptar únicamente los formatos permitidos, sin confiar en la extensión del nombre.

#### Scenario: JPEG o PNG válido

- **WHEN** el vendedor adjunta un archivo cuyo contenido real es JPEG o PNG y cumple los límites
- **THEN** el sistema acepta la fotografía

#### Scenario: Extensión falsa

- **WHEN** el vendedor adjunta un archivo con extensión de imagen pero contenido no permitido
- **THEN** el sistema rechaza la fotografía

#### Scenario: Archivo no soportado

- **WHEN** el vendedor adjunta un archivo de un formato no permitido
- **THEN** el sistema lo rechaza e informa los formatos aceptados

### Requirement: Límites de tamaño y cantidad

El sistema SHALL imponer un tamaño máximo por fotografía, una cantidad máxima de fotografías por publicación y dimensiones máximas, rechazando los archivos que excedan los límites.

#### Scenario: Fotografía demasiado grande

- **WHEN** el vendedor adjunta una fotografía que supera el tamaño máximo permitido
- **THEN** el sistema rechaza la publicación e indica el límite

#### Scenario: Demasiadas fotografías

- **WHEN** el vendedor adjunta más fotografías que el máximo permitido
- **THEN** el sistema rechaza el exceso e informa la cantidad permitida

#### Scenario: Imagen con dimensiones excesivas

- **WHEN** el vendedor adjunta una imagen que supera las dimensiones máximas
- **THEN** el sistema la rechaza

### Requirement: Consultas de catálogo sin blobs

El sistema SHALL consultar únicamente los metadatos de las fotografías al listar el catálogo y SHALL NOT cargar todos los contenidos binarios en memoria para renderizar los listados.

#### Scenario: Listado del catálogo

- **WHEN** se renderiza el catálogo con publicaciones que tienen fotografías
- **THEN** el sistema obtiene los metadatos necesarios y las imágenes se solicitan por su endpoint individual

### Requirement: Servicio de fotografías

El sistema SHALL servir cada fotografía con su tipo de contenido, longitud y validador de caché, y SHALL permitir solicitudes condicionales sin volver a transferir el contenido sin cambios.

#### Scenario: Descarga de fotografía

- **WHEN** un cliente solicita una fotografía pública
- **THEN** el sistema responde con el contenido, el tipo de contenido correcto y encabezados de caché

#### Scenario: Solicitud condicional

- **WHEN** un cliente solicita una fotografía con un validador que coincide con la versión actual
- **THEN** el sistema responde que el contenido no ha cambiado y no reenvía el cuerpo

### Requirement: Nombres seguros y metadatos

El sistema SHALL generar identificadores propios para el almacenamiento y SHALL NOT usar el nombre original del archivo como ruta o clave de persistencia.

#### Scenario: Nombre original peligroso

- **WHEN** el vendedor adjunta una fotografía con un nombre que contiene rutas o caracteres especiales
- **THEN** el sistema la guarda con un identificador seguro y no utiliza ese nombre como clave

### Requirement: Asociación íntegra con la publicación

Cada fotografía SHALL pertenecer a una publicación existente y SHALL conservarse mientras la publicación mantenga su registro, incluso si cambia de estado.

#### Scenario: Retirada de publicación

- **WHEN** una publicación es retirada por el vendedor o el administrador
- **THEN** el registro de sus fotografías se conserva y deja de servirse públicamente
