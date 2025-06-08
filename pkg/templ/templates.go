package templ

import (
	"github.com/labstack/echo/v4"
	"html/template"
	"io"
	"os"
	"path/filepath"
)

type Templates struct {
	template *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.template.ExecuteTemplate(w, name, data)
}

func NewTemplate() *Templates {
	var files []string

	err := filepath.Walk("views", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".gohtml" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	tmpl := template.Must(template.ParseFiles(files...))

	return &Templates{
		template: tmpl,
	}
}
