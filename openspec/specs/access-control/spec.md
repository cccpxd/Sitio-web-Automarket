# access-control Specification

## Purpose
Define los roles de la plataforma y las reglas de autorización que protegen cada operación, incluyendo la propiedad de las publicaciones y la ventana temporal de edición.

## Requirements

### Requirement: Acceso público sin sesión

Los visitantes SHALL poder consultar el catálogo, el detalle de una publicación activa, las fotografías públicas y usar el buscador y los filtros sin iniciar sesión.

#### Scenario: Visitante consulta el catálogo

- **WHEN** un visitante sin sesión abre el catálogo o el detalle de una publicación activa
- **THEN** el sistema le muestra la información pública sin pedir autenticación

#### Scenario: Visitante intenta publicar

- **WHEN** un visitante sin sesión intenta acceder al formulario de publicación
- **THEN** el sistema le exige iniciar sesión antes de continuar

### Requirement: Permisos por rol

El sistema SHALL permitir que solo los vendedores autenticados publiquen y gestionen sus propios vehículos, y que solo los administradores retiren publicaciones y consulten el registro de moderación.

#### Scenario: Vendedor publica

- **WHEN** un vendedor autenticado envía el formulario de publicación con datos válidos
- **THEN** el sistema crea la publicación asociada a su cuenta

#### Scenario: Vendedor intenta moderar

- **WHEN** un vendedor autenticado intenta retirar una publicación o abrir el registro de moderación
- **THEN** el sistema rechaza la operación con un error de permisos

#### Scenario: Administrador modera

- **WHEN** un administrador autenticado retira una publicación indicando una razón
- **THEN** el sistema aplica la retirada y registra la acción

### Requirement: Verificación de propiedad

El sistema SHALL verificar en el backend que un vendedor solo pueda editar o retirar publicaciones cuyo `seller_id` coincida con su cuenta, incluso si la petición se envía directamente al endpoint.

#### Scenario: Vendedor edita publicación ajena

- **WHEN** un vendedor envía una edición sobre una publicación de otro vendedor
- **THEN** el sistema rechaza la operación con un error de permisos y no modifica los datos

#### Scenario: Vendedor retira publicación ajena

- **WHEN** un vendedor intenta retirar una publicación que no le pertenece
- **THEN** el sistema rechaza la operación y la publicación permanece en su estado anterior

### Requirement: Ventana de edición de diez minutos

El sistema SHALL permitir al vendedor editar su publicación únicamente dentro de los primeros diez minutos posteriores a su creación, validando el plazo en el servidor.

#### Scenario: Edición dentro del plazo

- **WHEN** el vendedor envía una edición antes de que transcurran diez minutos desde la creación
- **THEN** el sistema aplica los cambios y registra la fecha de edición

#### Scenario: Edición fuera del plazo

- **WHEN** el vendedor envía una edición después de transcurridos diez minutos
- **THEN** el sistema la rechaza con un mensaje que indica que el plazo de edición expiró y no modifica la publicación

### Requirement: Acceso a conversaciones

El sistema SHALL permitir leer y responder una conversación únicamente al comprador y al vendedor que participan en ella; el administrador SHALL tener acceso de solo lectura para fines de moderación.

#### Scenario: Tercero ajeno intenta leer una conversación

- **WHEN** un usuario autenticado que no participa en la conversación solicita su contenido
- **THEN** el sistema rechaza la solicitud sin revelar los mensajes

#### Scenario: Participante responde

- **WHEN** el comprador o el vendedor participante envía un mensaje en su conversación
- **THEN** el sistema acepta el mensaje y lo muestra a la otra parte

### Requirement: Protección de datos sensibles

El sistema SHALL excluir contraseñas, hashes, tokens de sesión y datos privados de todas las respuestas y vistas públicas, exponiendo del vendedor solo la información apropiada para el contacto.

#### Scenario: Respuesta pública del vendedor

- **WHEN** un visitante consulta el detalle de una publicación
- **THEN** la respuesta muestra el nombre público del vendedor y no incluye su email, contraseña, hash ni token de sesión

#### Scenario: Intento de acceso directo sin sesión

- **WHEN** se invoca directamente un endpoint protegido sin una sesión válida
- **THEN** el sistema responde con un error de autenticación y no ejecuta la operación
