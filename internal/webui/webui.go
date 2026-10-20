package webui

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
)

//go:embed templates/*.html
var templateFiles embed.FS

//go:embed static
var staticFiles embed.FS

// Renderer executes embedded HTML templates. Embedding makes the binary
// self-contained, which is required in Vercel's read-only runtime filesystem.
type Renderer struct {
	templates *template.Template
}

func NewRenderer() (*Renderer, error) {
	templates, err := template.ParseFS(templateFiles, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse embedded templates: %w", err)
	}
	return &Renderer{templates: templates}, nil
}

func (r *Renderer) Render(w io.Writer, name string, data any) error {
	if r.templates.Lookup(name) == nil {
		return fmt.Errorf("template %q not found", name)
	}
	if err := r.templates.ExecuteTemplate(w, name, data); err != nil {
		return fmt.Errorf("execute template %q: %w", name, err)
	}
	return nil
}

// StaticFS returns the embedded static assets rooted at the /static URL.
func StaticFS() (fs.FS, error) {
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return nil, fmt.Errorf("build static filesystem: %w", err)
	}
	return sub, nil
}
