# Spec Delta

## MODIFIED Requirements

### Requirement: Filtros combinables

El sistema SHALL permitir filtrar por ciudad, año, precio mínimo y precio máximo en pesos colombianos (COP), y SHALL poder combinar estos filtros con la búsqueda de texto.

#### Scenario: Filtros combinados

- **WHEN** el usuario busca por marca, ciudad, año y rango de precio en COP a la vez
- **THEN** el sistema devuelve únicamente publicaciones que cumplen todos los criterios

#### Scenario: Rango de precio inválido

- **WHEN** el usuario envía un precio mínimo mayor que el máximo, un valor no numérico o un valor con decimales
- **THEN** el sistema informa el error y no ejecuta un filtro inconsistente

## ADDED Requirements

### Requirement: Filtro por ciudad

El sistema SHALL filtrar las publicaciones activas por ciudad, comparando el texto normalizado sin distinguir mayúsculas ni espacios sobrantes, y SHALL NOT concatenar el valor introducido en el texto SQL.

#### Scenario: Ciudad con coincidencias

- **WHEN** el usuario filtra por `  bogotá  `
- **THEN** el sistema devuelve únicamente publicaciones activas cuya ciudad coincide con `Bogotá` sin distinguir mayúsculas

#### Scenario: Ciudad sin coincidencias

- **WHEN** el usuario filtra por una ciudad que no tiene publicaciones activas
- **THEN** el sistema muestra el estado sin resultados y permite limpiar los filtros

#### Scenario: Ciudad combinada con búsqueda de texto

- **WHEN** el usuario combina una ciudad con una marca o un año
- **THEN** el sistema devuelve solo las publicaciones que cumplen la ciudad y el resto de criterios

#### Scenario: Publicación sin ciudad y filtro activo

- **WHEN** el usuario aplica un filtro de ciudad y existen publicaciones activas sin ciudad registrada
- **THEN** el sistema excluye esas publicaciones de los resultados

#### Scenario: Intento de inyección en el filtro de ciudad

- **WHEN** el usuario introduce caracteres propios de una inyección SQL en el filtro de ciudad
- **THEN** el sistema los trata como texto de búsqueda y no altera la consulta
