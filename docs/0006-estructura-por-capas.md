# ADR-0006: Estructura del código por capas

- **Estado:** Aceptado
- **Fecha:** 2026-09-27
- **Autores:** [nombres del equipo]

## Contexto

Somos un equipo de 3 personas trabajando en paralelo. Necesitamos que el código sea fácil de entender, de probar y de repartir. El curso pide separación estructural: entrada, lógica y salida en lugares distintos.

## Decisión

Organizamos el código en cuatro capas.

- **`handlers`:** reciben la petición HTTP, leen los datos, llaman al servicio y responden. No tienen reglas de negocio.
- **`servicios`:** contienen las reglas de negocio. Reciben datos por parámetros y devuelven resultados. No conocen Gin.
- **`repositorios`:** guardan y leen datos (ADR-0001). Se usan por medio de interfaces.
- **`modelos`:** las estructuras de datos (Usuario, Vehículo, Sesión, etc.).

Reglas:
- Las dependencias van en un solo sentido: `handlers` → `servicios` → `repositorios`.
- No hay variables globales. Todo se pasa por parámetros y constructores.
- `main` es corto: crea las piezas, las conecta y arranca el servidor.

## Alternativas consideradas

1. **Todo en un paquete.** Es rápido al inicio, pero se enreda pronto. Descartada.
2. **Estructura por funcionalidad** (una carpeta por usuarios, vehículos, etc.). Escala bien, pero cuesta más armarla al empezar. Descartada por ahora.

## Consecuencias

- **A favor:** la lógica se prueba sin levantar el servidor. Cada persona puede trabajar en una capa sin pisar a las otras.
- **En contra:** hay más archivos y más código de "conexión" entre capas.
