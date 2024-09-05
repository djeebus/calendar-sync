package templates

import (
	"embed"
	"github.com/labstack/echo/v4"
	"html/template"
	"io"
)

type Templates struct {
	templates *template.Template
}

//go:embed html
var templateFS embed.FS

func New() *Templates {
	return &Templates{templates: template.Must(template.ParseFS(templateFS, "html/*.html"))}
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
