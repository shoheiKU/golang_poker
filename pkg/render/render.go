package render

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/shoheiKU/web_application/GO_UDEMY/pkg/config"
	"github.com/shoheiKU/web_application/GO_UDEMY/pkg/models"
)

const mainTmplDir = "./templates/"

var app *config.AppConfig
var functions = template.FuncMap{}

// NewTemplates set the config for the template package
func NewTemplates(a *config.AppConfig) {
	app = a
}

// RenderTemplate renders a template.
func RenderTemplate(w http.ResponseWriter, tmpl string, td *models.TemplateData) {
	var tc map[string]*template.Template
	if app.UseCache {
		tc = app.TemplateCache
	} else {
		tc = CreateTemplateCache(mainTmplDir)
	}
	err := tc[tmpl].Execute(w, td)
	if err != nil {
		fmt.Println("error executing tmpl", err)
	}
}

// CreateTemplateCache creates a template cache as a map
func CreateTemplateCache(dir string) map[string]*template.Template {
	myCache := map[string]*template.Template{}
	pages, err := filepath.Glob(dir + "*.page.tmpl")
	if err != nil {
		fmt.Println("error globbing pages", err)
	}
	for _, page := range pages {
		tmpl_name := filepath.Base(page)
		t, err := template.New(tmpl_name).Funcs(functions).ParseFiles(dir + tmpl_name)
		if err != nil {
			fmt.Println("error parsing a file", err)
		}
		layouts, err := filepath.Glob(dir + "*.layout.tmpl")
		if err != nil {
			fmt.Println("error globbing pages", err)
		}
		for _, layout := range layouts {
			layout_name := filepath.Base(layout)
			t, err = t.ParseGlob(dir + layout_name)
			if err != nil {
				fmt.Println("error parsing a file", err)
			}
		}
		myCache[tmpl_name] = t
	}
	return myCache
}
