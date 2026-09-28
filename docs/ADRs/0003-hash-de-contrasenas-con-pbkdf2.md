# ADR-0003: Proteger las contraseñas con PBKDF2

- **Estado:** Aceptado
- **Fecha:** 2026-09-27
- **Autores:** [nombres del equipo]

## Contexto

Las contraseñas no se pueden guardar en texto plano. El enunciado prohíbe `bcrypt` y cualquier librería externa (esto incluye `golang.org/x/crypto`). Necesitamos un método lento a propósito, para que sea caro probar contraseñas por fuerza bruta.

## Decisión

Usamos PBKDF2 con HMAC-SHA-256 del paquete `crypto/pbkdf2` de la librería estándar.

- Cada usuario tiene una sal única de 16 bytes o más, generada con `crypto/rand`.
- Se usa un número alto de iteraciones (por ejemplo, 600.000). Se guarda con el hash para poder subirlo en el futuro.
- Se guarda un solo texto por usuario, con formato `iteraciones$sal$hash`.
- Al iniciar sesión, se calcula el hash con los mismos datos y se compara con `crypto/subtle.ConstantTimeCompare`.
- Las contraseñas y los hashes no se escriben en logs.

## Alternativas consideradas

1. **bcrypt.** Es lo que se usó en CanchaYA. Prohibido por el enunciado.
2. **SHA-256 simple.** Es muy rápido, y por eso débil frente a fuerza bruta. Descartada.
3. **SHA-256 con sal.** Mejor, pero sigue siendo rápido de atacar. Descartada.
4. **Hash iterado hecho a mano.** Es el plan B si la versión de Go no incluye `crypto/pbkdf2`.

## Consecuencias

- **A favor:** sin dependencias externas. Es un método reconocido y lento a propósito.
- **En contra:** `crypto/pbkdf2` está en la librería estándar desde Go 1.24. Hay que confirmar la versión con `go version`. Con muchas iteraciones, el login tarda un poco más y usa CPU.
