# ADR-0009: Borrado lógico de publicaciones

- **Estado:** Aceptado
- **Fecha:** 2026-11-06

## Contexto

Retirar una publicación no debe borrar conversaciones, fotografías ni evidencia de moderación. El catálogo público debe ocultarla inmediatamente.

## Decisión

Mantener la fila de `cars` y cambiar `status` a `withdrawn` para el vendedor o `removed` para moderación. Las consultas públicas filtran `status = 'active'`; las consultas propias y administrativas conservan los estados restantes.

## Alternativas y consecuencias

El borrado físico rompe claves foráneas y elimina evidencia. Un registro separado complica el modelo. El estado preserva integridad y auditoría, a cambio de exigir filtros explícitos en cada consulta pública.
