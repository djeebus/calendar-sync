package www

import (
	"embed"
	"github.com/labstack/echo/v4"
	"html/template"
	"io"
)

type templates struct {
	templates *template.Template
}

//go:embed templates
var templateFS embed.FS

func newTemplates() *templates {
	return &templates{templates: template.Must(template.ParseFS(templateFS, "templates/*.html"))}
}

func (t *templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
