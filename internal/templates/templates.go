package templates

import (
	"forum/internal/models"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

var templates map[string]*template.Template

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
		basePath := filepath.Join("templates", "base.html")
		pagePath := filepath.Join("templates", page+".html")

		tmpl, err := template.New("base").Funcs(funcs).ParseFiles(basePath, pagePath)
		if err != nil {
			return err
		}
		templates[page] = tmpl
	}

	return nil
}

func Render(w http.ResponseWriter, name string, data models.TemplateData, status int) {
	tmpl, ok := templates[name]
	if !ok {
		http.Error(w, "Template missing", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if status > 0 {
		w.WriteHeader(status)
	}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("render: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
