# ADR-0001: Guardar los datos en memoria y en archivos JSON

- **Estado:** Aceptado
- **Fecha:** 2026-09-27
- **Autores:** [nombres del equipo]

## Contexto

AutoMarket necesita guardar usuarios, vehículos y sesiones. El enunciado solo permite la librería estándar de Go y Gin. Por eso no podemos usar un ORM ni un driver de base de datos (SQLite, MySQL, etc.).

Gin atiende varias peticiones al mismo tiempo. El acceso a los datos debe ser seguro para varios hilos.

## Decisión

Los datos viven en memoria (mapas y slices de Go) y se guardan en archivos JSON. Al arrancar, el servidor carga los archivos. Cada cambio se escribe de vuelta al archivo.

- El acceso a los datos se protege con `sync.Mutex` (o `sync.RWMutex`).
- El acceso a los datos va detrás de interfaces pequeñas, por ejemplo `UsuarioRepositorio` y `VehiculoRepositorio`.
- Al escribir, se crea un archivo temporal y luego se renombra. Así un fallo a medias no daña el archivo original.

## Alternativas consideradas

1. **Solo memoria.** Es lo más simple, pero todo se pierde al apagar el servidor. Descartada.
2. **Un archivo JSON por registro.** Es fácil de revisar a mano, pero se vuelve lento y desordenado cuando crece. Descartada.
3. **Base de datos con ORM o driver externo.** Prohibida por el enunciado.

## Consecuencias

- **A favor:** no se pierden datos al reiniciar. La lógica no depende del almacenamiento: cambiarlo solo toca los repositorios.
- **En contra:** el rendimiento baja si hay muchos datos, porque se reescribe el archivo. No hay consultas complejas. Hay que manejar bien el bloqueo para evitar condiciones de carrera.
