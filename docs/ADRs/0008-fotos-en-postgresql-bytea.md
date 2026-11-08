# ADR-0008: Fotografías en PostgreSQL como BYTEA

- **Estado:** Aceptado
- **Fecha:** 2026-11-06

## Contexto

Las fotografías deben permanecer asociadas a cada publicación y estar disponibles para usuarios autorizados. El proyecto ya depende de PostgreSQL y el volumen esperado es pequeño.

## Decisión

Guardar cada imagen normalizada como JPEG en `car_photos.data BYTEA`, junto con tipo, tamaño, hash SHA-256, dimensiones, posición y nombre original sanitizado. Los listados consultan solo metadatos; `/photos/:id` consulta una fila y responde `ETag`, `Content-Length` y caché.

## Alternativas

- **Large Objects:** permiten streaming, pero requieren lifecycle y permisos adicionales.
- **Almacenamiento externo:** escala mejor, pero añade credenciales, proveedor y un segundo sistema de copias.
- **BYTEA:** simplifica transacciones y copias; a cambio, aumenta el tamaño de la base y no es ideal para un catálogo masivo.

## Consecuencias

Cada archivo está limitado a 5 MiB, hasta seis fotos y 6000 px. El tipo se detecta por firma y la imagen se decodifica antes de persistir; la extensión y el nombre original no son confiables.
