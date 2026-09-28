# AutoMarket

Proyecto web para una tienda de autos usados, desarrollado en Go con Gin.

## Descripción

Esta rama contiene una versión inicial de la aplicación con una landing page de AutoMarket, una interfaz visual para mostrar vehículos recomendados y una estructura básica preparada para crecer con más funcionalidades.

## Tecnologías

- Go
- Gin Framework
- HTML
- CSS
- Templates de servidor

## Estructura principal

- `main.go` — punto de entrada de la aplicación
- `internal/router/router.go` — configuración de rutas
- `internal/handlers/home.go` — controlador de la página principal
- `templates/index.html` — vista de la página de inicio
- `static/css/style.css` — estilos de la interfaz


## Cómo ejecutar

1. Instala Go en tu sistema.
2. En la raíz del proyecto ejecuta:

```bash
go run .
```

3. Abre en tu navegador:

```text
http://localhost:8080
```

## Nota

La aplicación está en una etapa inicial y sirve como base para continuar con el desarrollo del sitio web, módulos de catálogo y más funcionalidades del negocio.
