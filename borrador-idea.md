# Prompt para OpenCode: finalizar la plataforma de venta de vehículos

Usa el flujo `/opsx-propose` para analizar el proyecto y preparar una propuesta de implementación. El objetivo es terminar la página web existente para convertirla en una plataforma donde los vendedores publiquen vehículos y los visitantes puedan consultarlos sin iniciar sesión.

Antes de proponer cambios:
- Inspecciona la estructura actual del proyecto, el framework, el frontend, el backend, la autenticación y el esquema de PostgreSQL.
- Identifica qué funcionalidades ya existen y cuáles faltan.
- Respeta la arquitectura, las convenciones y los estilos actuales del proyecto cuando sea razonable.
- No reemplaces tecnologías ni reescribas módulos que ya funcionan sin una razón técnica clara.
- Si falta información, señala los supuestos en la propuesta. No inventes funcionalidades ajenas a este alcance.

## 1. Tipos de usuario y permisos

La plataforma tiene tres tipos de usuario: visitante, vendedor y administrador.

### Visitante
- Puede entrar al sitio y consultar todas las publicaciones activas sin iniciar sesión.
- Puede usar el buscador y los filtros.
- Puede abrir el detalle de cada vehículo y ver sus fotografías, precio, año, modelo y descripción.
- Puede iniciar el contacto con el vendedor mediante el botón de interés. Si el sistema exige autenticación para enviar mensajes, solicita iniciar sesión o registrarse y conserva el mensaje que quería enviar.

### Vendedor
- Puede consultar las mismas publicaciones activas que ve un visitante.
- Tiene un apartado llamado **Ventas** que muestra los vehículos publicados en la plataforma. Este apartado debe usar la misma fuente de datos que el catálogo público, para evitar que existan listados distintos o desactualizados.
- Puede crear publicaciones de vehículos.
- Puede consultar y gestionar sus propias publicaciones, según las funciones que ya soporte el proyecto.
- Puede recibir y responder mensajes de interesados.
- Puede usar el buscador y los filtros.

### Administrador
- Puede consultar las publicaciones.
- Puede eliminar cualquier publicación que incumpla las reglas de la plataforma.
- Al eliminar una publicación, debe registrar una razón obligatoria.
- La eliminación no requiere aprobación previa. Cuando un vendedor publica un vehículo, este aparece de inmediato para los visitantes y los demás vendedores.
- Puede consultar un registro de moderación con la publicación afectada, el administrador responsable, la razón y la fecha de eliminación.

Aplica controles de acceso tanto en el frontend como en el backend. No confíes únicamente en ocultar botones o apartados.

## 2. Catálogo público de vehículos

Crea o termina un catálogo de vehículos en venta.

Cada publicación debe mostrar:
- Fotografías del vehículo.
- Modelo o referencia.
- Marca, si el modelo de datos actual la contempla o si hace falta para mejorar la búsqueda.
- Año.
- Precio.
- Descripción.
- Información pública del vendedor que sea apropiada para el contacto.
- Fecha de publicación, si encaja con el diseño actual.

Requisitos:
- Cualquier visitante puede consultar el catálogo sin iniciar sesión.
- Las publicaciones nuevas aparecen para todos los visitantes y vendedores después de guardarse correctamente.
- El catálogo debe mostrar únicamente publicaciones activas.
- No agregues vehículos, vendedores ni mensajes ficticios como datos de ejemplo.
- La plataforma debe iniciar con el catálogo vacío cuando no existan publicaciones reales.
- Incluye estados vacíos claros, por ejemplo: “Todavía no hay vehículos publicados”.

## 3. Publicar un vehículo

Crea un apartado para que los vendedores publiquen vehículos en venta.

El formulario debe incluir:
- Modelo o referencia del vehículo.
- Marca, si corresponde al modelo de datos.
- Año.
- Precio.
- Descripción.
- Una o varias fotografías.

Requisitos:
- Valida los campos obligatorios y los formatos.
- Valida que el precio y el año sean valores válidos.
- Limita los formatos y el tamaño de las imágenes.
- Muestra una vista previa de las fotografías antes de publicar.
- Permite quitar o cambiar fotografías antes de guardar.
- Al publicar correctamente, la publicación debe aparecer en el catálogo público y en el apartado Ventas de los vendedores.
- No debe existir un paso de aprobación del administrador.
- Evita publicaciones incompletas o duplicaciones causadas por varios clics.

## 4. Almacenamiento de fotografías en PostgreSQL

Investiga la mejor forma de guardar las fotografías teniendo en cuenta que la base de datos del proyecto es PostgreSQL.

Compara brevemente estas alternativas:
- Guardar los archivos como `BYTEA` en PostgreSQL.
- Usar Large Objects de PostgreSQL.
- Guardar las imágenes en un servicio de almacenamiento externo y conservar sus referencias en PostgreSQL.

La necesidad del proyecto es mantener las fotografías asociadas a cada publicación y que estén disponibles para todos los usuarios autorizados.

Elige una alternativa según la arquitectura y el tamaño esperado del proyecto. Explica sus ventajas, desventajas y consecuencias para copias de seguridad, rendimiento y mantenimiento. Si la decisión es guardar las imágenes directamente en PostgreSQL, diseña el esquema y la forma de servirlas sin cargar todas las imágenes en memoria al consultar el catálogo.

No guardes imágenes codificadas en Base64 dentro de campos de texto. Implementa validación del tipo real de archivo, tamaño máximo y nombres seguros. Evita confiar únicamente en la extensión del archivo.

## 5. Buscador y filtros

Implementa un buscador disponible para los tres tipos de usuario: visitante, vendedor y administrador.

Debe permitir buscar por:
- Marca o modelo.
- Año.
- Precio mínimo y máximo.
- Texto de la descripción, si la arquitectura actual permite hacerlo de forma eficiente.

Ejemplo de búsqueda: `Mazda 2013`.

Requisitos:
- La búsqueda por texto debe reconocer combinaciones de marca, modelo y año.
- Los filtros deben poder combinarse.
- Debe existir una opción para limpiar los filtros.
- Los resultados deben mostrar únicamente publicaciones activas.
- La búsqueda debe consultar los datos del backend y no limitarse a filtrar los vehículos que ya estén cargados en el navegador.
- Incluye paginación o carga progresiva si el volumen de publicaciones lo requiere.
- Valida y parametriza las consultas para evitar inyección SQL.

## 6. Mensajes entre interesados y vendedores

Agrega una función de mensajería asociada a las publicaciones.

En cada publicación debe existir un botón para contactar al vendedor, con un texto inicial predeterminado:

> Estoy interesado en tu carro: [descripción del carro]

Requisitos:
- Sustituye `[descripción del carro]` por la descripción real de la publicación. Si resulta demasiado extensa, usa una referencia breve del vehículo, como marca, modelo y año.
- El interesado debe revisar y enviar el mensaje.
- El mensaje debe quedar asociado a la publicación, al usuario remitente y al vendedor destinatario.
- El vendedor debe recibir el mensaje en un apartado **Mensajes**.
- El vendedor debe poder abrir una conversación y responder.
- El apartado Mensajes debe mostrar las conversaciones y permitir identificar la publicación relacionada.
- Define permisos para que cada usuario consulte únicamente las conversaciones en las que participa, mientras que el administrador conserva los permisos de moderación que correspondan.
- Si los visitantes anónimos no pueden mantener conversaciones, solicita iniciar sesión antes del envío y conserva el texto inicial para no obligarlos a escribirlo de nuevo.
- Evita que un usuario se envíe mensajes a sí mismo.
- Valida los mensajes en el backend y almacena fechas de envío y lectura si el diseño lo contempla.

## 7. Moderación de publicaciones

El administrador no aprueba publicaciones antes de su aparición. La moderación funciona después de publicar.

Implementa:
- Acción para eliminar una publicación.
- Campo obligatorio para indicar la razón.
- Confirmación antes de eliminar.
- Registro de auditoría con administrador, publicación, razón y fecha.
- La publicación eliminada debe dejar de aparecer en el catálogo público, en Ventas y en los resultados del buscador.
- Conserva el registro de moderación. No borres la evidencia de la acción administrativa.
- Define si conviene usar borrado lógico mediante un estado, en lugar de eliminar físicamente el registro. Prioriza conservar la integridad de mensajes y referencias.

## 8. Diseño visual

Actualiza el HTML y CSS para lograr una interfaz formal, limpia y consistente.

Condiciones:
- No uses azul como color principal ni como color de acento.
- Elige una paleta sobria. Puedes evaluar tonos carbón, blanco, gris y un acento cálido discreto.
- Mantén buen contraste y legibilidad.
- Usa una jerarquía visual clara para títulos, precios, datos técnicos y botones.
- Las tarjetas de vehículos deben dar prioridad a las fotografías y al precio.
- El diseño debe adaptarse a móviles, tabletas y escritorio.
- Mantén consistencia entre catálogo, detalle del vehículo, formulario de publicación, Ventas, Mensajes y administración.
- Usa etiquetas y mensajes de validación claros.
- Evita efectos visuales innecesarios.

## 9. Base de datos y seguridad

Revisa el esquema existente antes de crear tablas o columnas.

La propuesta debe contemplar, según corresponda:
- Publicaciones de vehículos.
- Fotografías y su relación con cada vehículo.
- Estados de publicación.
- Conversaciones y mensajes.
- Registro de moderación.
- Relaciones con usuarios y roles.

Requisitos técnicos:
- Crea migraciones compatibles con el sistema que ya usa el proyecto.
- Define claves foráneas, índices y restricciones.
- Usa transacciones cuando una operación afecte varias tablas.
- Aplica autorización en cada endpoint.
- Valida los datos en el servidor.
- Protege las rutas y operaciones administrativas.
- No expongas contraseñas, tokens ni datos privados.
- Evita consultas SQL construidas mediante concatenación de texto.
- Mantén la integridad de las conversaciones si una publicación se elimina.

## 10. Pruebas y criterios de aceptación

Incluye en la propuesta pruebas para comprobar estos escenarios:

1. Un visitante sin sesión consulta el catálogo.
2. Un visitante busca `Mazda 2013` y obtiene resultados compatibles.
3. Un vendedor publica un vehículo con fotografías y la publicación aparece en el catálogo público.
4. Otro vendedor ve esa misma publicación en Ventas.
5. El catálogo y Ventas muestran los mismos vehículos activos.
6. Un vendedor recibe y responde un mensaje asociado a su publicación.
7. Un usuario ajeno a una conversación no puede leerla.
8. Un administrador elimina una publicación e indica una razón.
9. La publicación eliminada deja de aparecer en el catálogo, Ventas y buscador.
10. El registro de moderación conserva la razón y los datos de la acción.
11. El catálogo vacío muestra un estado informativo y no contiene datos ficticios.
12. Los formularios rechazan imágenes inválidas y datos incorrectos.
13. Los permisos se cumplen aunque alguien intente llamar directamente a los endpoints.

## 11. Entregables de la propuesta

Usa `/opsx-propose` para generar una propuesta que incluya:

- Resumen de la solución.
- Análisis de la arquitectura existente.
- Funcionalidades que ya están implementadas y las que faltan.
- Cambios de frontend y backend.
- Cambios en el esquema de PostgreSQL y migraciones.
- Decisión justificada para el almacenamiento de fotografías.
- Endpoints o servicios necesarios.
- Reglas de permisos para los tres tipos de usuario.
- Flujo de publicación, búsqueda, moderación y mensajería.
- Pruebas necesarias.
- Orden recomendado de implementación y dependencias entre tareas.

No inventes ventas de ejemplo. No ejecutes una implementación destructiva ni elimines datos existentes durante la fase de propuesta. Si detectas decisiones que requieren confirmación, enuméralas claramente y recomienda una opción.

El resultado debe ajustarse al proyecto existente y dejar una ruta concreta para terminar la plataforma.
