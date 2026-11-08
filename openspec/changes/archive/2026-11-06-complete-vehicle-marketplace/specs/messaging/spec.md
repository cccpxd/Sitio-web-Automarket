# Spec Delta

## Purpose

Permite que un interesado contacte al vendedor de una publicación y mantenga una conversación privada asociada a ese vehículo.

## ADDED Requirements

### Requirement: Botón de interés en cada publicación

Cada publicación activa SHALL mostrar una acción para contactar al vendedor con un mensaje inicial predeterminado que sustituya la descripción real del vehículo o, si es demasiado extensa, una referencia breve con marca, modelo y año.

#### Scenario: Mensaje inicial con descripción

- **WHEN** un interesado abre la acción de contacto de una publicación con descripción breve
- **THEN** el sistema prepara el texto “Estoy interesado en tu carro: [descripción del carro]” con la descripción real

#### Scenario: Descripción demasiado extensa

- **WHEN** la descripción de la publicación es demasiado extensa para el mensaje inicial
- **THEN** el sistema usa una referencia breve con marca, modelo y año

### Requirement: Revisión y envío del mensaje

El interesado SHALL poder revisar y editar el mensaje inicial antes de enviarlo.

#### Scenario: Edición antes de enviar

- **WHEN** el interesado modifica el texto del mensaje inicial
- **THEN** el sistema envía el texto revisado y no el texto predeterminado

#### Scenario: Envío válido

- **WHEN** el interesado envía un mensaje con contenido válido
- **THEN** el sistema lo guarda y lo muestra en la conversación

### Requirement: Asociación del mensaje

Cada mensaje SHALL quedar asociado a la publicación, al usuario remitente y al vendedor destinatario.

#### Scenario: Mensaje asociado a la publicación

- **WHEN** se guarda un mensaje iniciado desde una publicación
- **THEN** el sistema conserva la referencia a esa publicación, al remitente y al vendedor

### Requirement: Bandeja y respuesta del vendedor

El vendedor SHALL recibir las conversaciones en un apartado Mensajes, SHALL poder identificar la publicación relacionada, abrir el hilo y responder.

#### Scenario: Vendedor recibe un mensaje

- **WHEN** un interesado envía un mensaje sobre una publicación del vendedor
- **THEN** la conversación aparece en el apartado Mensajes del vendedor con la publicación identificada

#### Scenario: Vendedor responde

- **WHEN** el vendedor abre una conversación y envía una respuesta
- **THEN** el mensaje se guarda y el interesado puede verlo en el mismo hilo

### Requirement: Permisos de conversación

El sistema SHALL permitir acceder a una conversación únicamente a sus participantes, mientras que el administrador SHALL conservar acceso de solo lectura para moderación.

#### Scenario: Usuario ajeno a la conversación

- **WHEN** un usuario que no participa intenta abrir o responder una conversación
- **THEN** el sistema rechaza el acceso sin revelar el contenido

#### Scenario: Administrador consulta una conversación

- **WHEN** un administrador abre una conversación para moderación
- **THEN** puede leer los mensajes pero no enviar respuestas como administrador

### Requirement: Autenticación para contactar

Si el interesado no tiene sesión, el sistema SHALL solicitarle iniciar sesión o registrarse antes del envío y SHALL conservar el texto que había escrito para no obligarlo a redactarlo de nuevo.

#### Scenario: Visitante anónimo contacta

- **WHEN** un visitante sin sesión pulsa contactar y escribe un mensaje
- **THEN** el sistema lo redirige a autenticarse y, al volver, restaura el texto para que pueda enviarlo

#### Scenario: Retorno tras autenticarse

- **WHEN** el visitante completa el inicio de sesión o el registro
- **THEN** regresa a la publicación de origen con el mensaje conservado

### Requirement: Prevención de auto-mensajes

El sistema SHALL impedir que un usuario inicie o envíe mensajes a su propia publicación.

#### Scenario: Vendedor intenta contactarse

- **WHEN** el vendedor propietario intenta enviar un mensaje sobre su propia publicación
- **THEN** el sistema rechaza la operación

### Requirement: Validación y trazabilidad de mensajes

El sistema SHALL validar el contenido en el servidor, limitar su longitud y almacenar la fecha de envío y, cuando corresponda, la fecha de lectura.

#### Scenario: Mensaje vacío o demasiado largo

- **WHEN** el usuario envía un mensaje vacío o superior al límite permitido
- **THEN** el sistema lo rechaza y muestra un error de validación

#### Scenario: Mensaje leído

- **WHEN** el destinatario abre una conversación con mensajes sin leer
- **THEN** el sistema registra la fecha de lectura de esos mensajes

### Requirement: Conservación de conversaciones

Las conversaciones y sus mensajes SHALL conservarse aunque la publicación sea retirada, y el hilo SHALL indicar que la publicación ya no está activa.

#### Scenario: Conversación sobre publicación retirada

- **WHEN** un participante abre una conversación cuya publicación fue retirada
- **THEN** puede leer el historial y ve un aviso de que la publicación ya no está disponible

#### Scenario: Nuevo contacto sobre publicación retirada

- **WHEN** un interesado intenta iniciar una conversación sobre una publicación retirada
- **THEN** el sistema no permite iniciar el contacto
