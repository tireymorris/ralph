package prompt

import (
	"embed"
	"strings"
	"text/template"
)

//go:embed templates/*.tmpl templates/partials/*.tmpl
var templateFS embed.FS

var templates *template.Template

func init() {
	templates = template.Must(
		template.New("prompts").Funcs(template.FuncMap{
			"add": func(a, b int) int { return a + b },
			"join": func(items []string, sep string) string {
				return strings.Join(items, sep)
			},
		}).ParseFS(templateFS, "templates/*.tmpl", "templates/partials/*.tmpl"),
	)
}
