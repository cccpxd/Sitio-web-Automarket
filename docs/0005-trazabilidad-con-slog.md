# ADR-0005: Trazabilidad con `log/slog` a archivo

- **Estado:** Aceptado
- **Fecha:** 2026-09-27
- **Autores:** [nombres del equipo]

## Contexto

AAA exige trazabilidad: poder saber quién hizo qué y cuándo. Solo podemos usar la librería estándar. `log/slog` ya viene con Go y permite logs con campos.

## Decisión

Usamos `log/slog` con salida a un archivo (por ejemplo, `logs/automarket.log`), en formato JSON.

- Un middleware registra las acciones importantes: inicio de sesión (bueno y fallido), cierre de sesión, y crear, editar o borrar vehículos y usuarios.
- Cada registro lleva: fecha, id de usuario, rol, acción, recurso afectado, resultado y dirección IP.
- **Nunca** se registran contraseñas, hashes ni tokens.
- El logger se crea en `main` y se pasa a quien lo necesite. No es una variable global.

## Alternativas consideradas

1. **`fmt.Println`.** No tiene formato ni campos, y se pierde al cerrar la consola. Descartada.
2. **Registro de auditoría separado del log técnico.** Es más completo, pero es más trabajo. Puede ser una mejora futura.

## Consecuencias

- **A favor:** sin dependencias. Los registros se pueden filtrar y leer fácil por ser JSON.
- **En contra:** el archivo crece sin límite. Hay que decidir cómo rotarlo. Quien pueda editar el archivo puede alterar la evidencia, así que hay que cuidar sus permisos.
