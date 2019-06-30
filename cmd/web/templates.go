package main

import (
	"html/template"
	"path/filepath"

	"github.com/rjpgt/battleship/pkg/forms"
	"github.com/rjpgt/battleship/pkg/models"
)

type templateData struct {
	Flash    string
	Form     *forms.Form
	GameID   string
	Opponent string
	Player   *models.Player
	Status   int
}

func newTemplateCache(dir string) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}
	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}
	return cache, nil
}
