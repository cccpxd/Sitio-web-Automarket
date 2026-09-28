# ADR-0002: Sesiones con token aleatorio guardado en el servidor

- **Estado:** Aceptado
- **Fecha:** 2026-09-27
- **Autores:** Cristian Camilo Colorado; Juan Jose Patiño; Juan Manuel Ramos

## Contexto

El sistema debe saber quién hace cada petición (autenticación, parte de AAA). El enunciado prohíbe librerías externas, incluido JWT. Además, el sistema usa HTTP sin TLS.

## Decisión

Al iniciar sesión correctamente, el servidor genera un token aleatorio y lo guarda con los datos de la sesión.

- El token se genera con `crypto/rand` (32 bytes, codificados en hexadecimal o base64).
- El servidor guarda: token, id de usuario, rol y fecha de vencimiento.
- El cliente envía el token en la cabecera `Authorization: Bearer <token>`.
- Un middleware de Gin busca el token, revisa que no haya vencido y deja el usuario y su rol en el contexto de la petición.
- Al cerrar sesión, el servidor borra el token.

## Alternativas consideradas

1. **JWT.** Es lo que se usó en el proyecto CanchaYA. Prohibido por el enunciado.
2. **Token firmado a mano con HMAC (JWT casero).** No hay que guardar sesiones, pero es fácil cometer errores de seguridad. No se puede cerrar una sesión antes de tiempo. Descartada.
3. **Cookie de sesión.** Es la misma idea que la escogida, pero el navegador guarda el token. Queda como posible mejora.

## Consecuencias

- **A favor:** se puede cerrar una sesión o bloquear a un usuario al instante. El token no dice nada por sí mismo, así que no filtra datos.
- **En contra:** el servidor debe guardar las sesiones. Si están solo en memoria, se pierden al reiniciar. Hay que limpiar las sesiones vencidas.
- **Riesgo aceptado:** sin TLS, el token viaja en texto plano y alguien en la red podría copiarlo. Es una limitación del enunciado. Por eso el token debe vencer pronto.
