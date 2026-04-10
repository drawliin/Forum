package templates

import (
	"bytes"
	"forum/internal/config"
	"forum/internal/models"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

var templates map[string]*template.Template

// InitTemplates parses the base layout together with each page template.
func InitTemplates() error {
	funcs := template.FuncMap{
		"formatUnix": func(ts int64) string {
			if ts == 0 {
				return ""
			}
			return time.Unix(ts, 0).Format("Jan 02, 2006 15:04")
		},
	}

	pages := []string{
		"home",
		"register",
		"login",
		"post_new",
		"post_view",
		"error",
	}

	templates = make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		basePath := config.ResolvePath(filepath.Join("templates", "base.html"))
		pagePath := config.ResolvePath(filepath.Join("templates", page+".html"))

		tmpl, err := template.New("base").Funcs(funcs).ParseFiles(basePath, pagePath)
		if err != nil {
			return err
		}
		templates[page] = tmpl
	}

	return nil
}

// Render executes one page template set and writes the final HTML response.
func Render(w http.ResponseWriter, name string, data models.TemplateData, status int) {
	tmpl, ok := templates[name]
	if !ok {
		http.Error(w, "Template missing", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if status != 0 {
		w.WriteHeader(status)
	}

	buff := bytes.Buffer{}
	err := tmpl.ExecuteTemplate(&buff, "base", data)
	if err != nil {
		log.Printf("render: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	} else {
		buff.WriteTo(w)
	}
}
