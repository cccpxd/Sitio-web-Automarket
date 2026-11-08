# ADR-0010: Precio entero en COP

- **Estado:** Aceptado
- **Fecha:** 2026-11-06

## Contexto

El formulario y el catálogo trabajan con pesos colombianos sin centavos visibles. Los filtros deben ser comparables en SQL y no depender de texto con separadores.

## Decisión

El formulario acepta solo pesos enteros positivos con separador de miles opcional. El dominio normaliza a `price_cents` como unidad mínima del peso (el valor ingresado multiplicado por 100), y la presentación elimina decimales y usa puntos como separadores de miles. La columna textual `price` se conserva temporalmente para compatibilidad y backfill.

## Alternativas y consecuencias

Guardar el precio como texto impide índices y rangos fiables. `NUMERIC` es legible, pero menos directo para contratos existentes. `BIGINT` evita errores de coma flotante y hace eficientes los filtros, a cambio de requerir conversión explícita en la frontera HTTP.
