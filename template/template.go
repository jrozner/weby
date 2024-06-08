package template

import (
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

var (
	ErrTemplateExists = errors.New("template already exists")
	ErrNoTemplate     = errors.New("template not found")
)

// GenerateTemplates produces a map of templates based on a common base layout. It expects the following directory
// structure
//
// templates/contacts/index.html
// templates/contacts/new.html
// templates/names/index.html
// templates/names/new.html
//
// The base directory name does not matter, but it must not include an underscore as the first character. The base
// layouts should occur within a directory called _layout in the root templates directory
func GenerateTemplates(templatesPath string) (map[string]*template.Template, error) {
	templates := make(map[string]*template.Template)

	base := template.Must(template.ParseGlob(path.Join(templatesPath, "_layouts", "*.html")))

	err := fs.WalkDir(os.DirFS(templatesPath), ".", func(p string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if strings.HasPrefix(p, "_") {
			return nil
		}

		if _, ok := templates[p]; ok {
			// this shouldn't be possible because the naming it based on file paths, but we should handle it if it does
			return ErrTemplateExists
		}

		cloned, err := base.Clone()
		if err != nil {
			return err
		}

		t, err := cloned.ParseFiles(path.Join(templatesPath, p))
		if err != nil {
			return err
		}

		templates[p] = t

		return nil
	})

	return templates, err
}

// Render finds the specified template in the map of templates, executes it, and writes it to the writer.
func Render(w http.ResponseWriter, status int, templates map[string]*template.Template, name string, data interface{}) error {
	t, ok := templates[name]
	if !ok {
		return ErrNoTemplate
	}

	name = path.Base(name)

	w.WriteHeader(status)
	return t.ExecuteTemplate(w, name, data)
}
