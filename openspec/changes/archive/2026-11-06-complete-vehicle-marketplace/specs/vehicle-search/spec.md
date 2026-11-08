# Spec Delta

## Purpose

Permite a visitantes, vendedores y administradores encontrar publicaciones activas mediante búsqueda de texto y filtros combinables ejecutados en el backend.

## ADDED Requirements

### Requirement: Búsqueda disponible para todos los roles

El sistema SHALL ofrecer el buscador y los filtros a visitantes sin sesión, vendedores y administradores.

#### Scenario: Visitante busca sin sesión

- **WHEN** un visitante sin sesión introduce un criterio de búsqueda
- **THEN** el sistema ejecuta la búsqueda y muestra resultados sin pedir autenticación

#### Scenario: Vendedor busca

- **WHEN** un vendedor autenticado usa el buscador desde Ventas o el catálogo
- **THEN** el sistema ejecuta la misma búsqueda sobre publicaciones activas

### Requirement: Búsqueda por texto combinado

El sistema SHALL interpretar el texto de búsqueda contra marca, modelo, año y descripción, reconociendo combinaciones como `Mazda 2013`.

#### Scenario: Búsqueda marca y año

- **WHEN** el usuario busca `Mazda 2013`
- **THEN** el sistema devuelve publicaciones cuya marca o modelo sea Mazda y cuyo año sea 2013, o que coincidan razonablemente con ambos términos

#### Scenario: Búsqueda solo por descripción

- **WHEN** el usuario busca una palabra presente en la descripción de una publicación activa
- **THEN** el sistema incluye esa publicación en los resultados

#### Scenario: Búsqueda sin coincidencias

- **WHEN** el texto no coincide con ninguna publicación activa
- **THEN** el sistema muestra un estado sin resultados y permite limpiar los filtros

### Requirement: Filtros combinables

El sistema SHALL permitir filtrar por año, precio mínimo y precio máximo en pesos colombianos (COP), y SHALL poder combinar estos filtros con la búsqueda de texto.

#### Scenario: Filtros combinados

- **WHEN** el usuario busca por marca, año y rango de precio en COP a la vez
- **THEN** el sistema devuelve únicamente publicaciones que cumplen todos los criterios

#### Scenario: Rango de precio inválido

- **WHEN** el usuario envía un precio mínimo mayor que el máximo, un valor no numérico o un valor con decimales
- **THEN** el sistema informa el error y no ejecuta un filtro inconsistente

### Requirement: Limpieza de filtros

El sistema SHALL ofrecer una acción para limpiar todos los criterios y volver al catálogo completo de publicaciones activas.

#### Scenario: Limpiar criterios

- **WHEN** el usuario activa la opción de limpiar filtros
- **THEN** el sistema descarta el texto y los filtros y muestra el catálogo activo sin criterios

### Requirement: Resultados solo activos

El sistema SHALL incluir en los resultados únicamente publicaciones con estado activo.

#### Scenario: Publicación retirada en la búsqueda

- **WHEN** una publicación retirada coincide con el texto o los filtros
- **THEN** el sistema la excluye de los resultados

### Requirement: Búsqueda en el backend

El sistema SHALL ejecutar la búsqueda y los filtros contra los datos del backend y SHALL NOT limitarse a filtrar los vehículos ya cargados en el navegador.

#### Scenario: Búsqueda con catálogo no cargado

- **WHEN** el usuario aplica un criterio desde una página que no contiene todas las publicaciones
- **THEN** el sistema consulta el backend y devuelve el conjunto completo que coincide

### Requirement: Paginación de resultados

El sistema SHALL paginar los resultados o usar carga progresiva cuando el volumen de publicaciones lo requiera, manteniendo los criterios aplicados.

#### Scenario: Resultados en varias páginas

- **WHEN** el número de coincidencias supera el tamaño de página
- **THEN** el sistema muestra controles de paginación y conserva los filtros al cambiar de página

### Requirement: Consultas parametrizadas

El sistema SHALL construir las consultas de búsqueda con parámetros y SHALL NOT concatenar valores introducidos por el usuario en el texto SQL.

#### Scenario: Intento de inyección SQL

- **WHEN** el usuario introduce caracteres propios de una inyección SQL en el buscador
- **THEN** el sistema los trata como texto de búsqueda y no altera la consulta
