package template

import (
	"html/template"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
)

// writeTree lays down a minimal template tree under a temp dir: a base layout
// that blocks "body", and a page whose body calls a "humanize" func.
func writeTree(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	layouts := path.Join(dir, "_layouts")
	if err := os.MkdirAll(layouts, 0o755); err != nil {
		t.Fatalf("mkdir layouts: %v", err)
	}

	pages := path.Join(dir, "pages")
	if err := os.MkdirAll(pages, 0o755); err != nil {
		t.Fatalf("mkdir pages: %v", err)
	}

	if err := os.WriteFile(path.Join(layouts, "base.html"), []byte(`{{define "base"}}{{block "body" .}}{{end}}{{end}}`), 0o644); err != nil {
		t.Fatalf("write base.html: %v", err)
	}

	if err := os.WriteFile(path.Join(pages, "index.html"), []byte(`{{define "body"}}{{humanize .Value}}{{end}}`), 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}

	return dir
}

// TestGenerateTemplatesFuncMap verifies that a custom FuncMap is registered on
// the base layout so pages that reference the func parse and execute correctly.
func TestGenerateTemplatesFuncMap(t *testing.T) {
	dir := writeTree(t)

	funcs := template.FuncMap{
		"humanize": func(s string) string { return "HUMAN:" + s },
	}

	templates, err := GenerateTemplates(dir, funcs)
	if err != nil {
		t.Fatalf("GenerateTemplates: %v", err)
	}

	key := "pages/index.html"
	if _, ok := templates[key]; !ok {
		t.Fatalf("expected template map to contain key %q, got keys %v", key, keysOf(templates))
	}

	rec := httptest.NewRecorder()
	data := struct{ Value string }{Value: "world"}
	if err := Render(rec, 200, templates, key, "base", data); err != nil {
		t.Fatalf("Render: %v", err)
	}

	if got := rec.Body.String(); !strings.Contains(got, "HUMAN:world") {
		t.Fatalf("expected output to contain %q, got %q", "HUMAN:world", got)
	}
}

// TestGenerateTemplatesNoFuncs verifies the backward-compatible no-funcs call
// still succeeds and produces the expected page key.
func TestGenerateTemplatesNoFuncs(t *testing.T) {
	dir := t.TempDir()

	layouts := path.Join(dir, "_layouts")
	if err := os.MkdirAll(layouts, 0o755); err != nil {
		t.Fatalf("mkdir layouts: %v", err)
	}

	pages := path.Join(dir, "pages")
	if err := os.MkdirAll(pages, 0o755); err != nil {
		t.Fatalf("mkdir pages: %v", err)
	}

	if err := os.WriteFile(path.Join(layouts, "base.html"), []byte(`{{define "base"}}{{block "body" .}}{{end}}{{end}}`), 0o644); err != nil {
		t.Fatalf("write base.html: %v", err)
	}

	// A page that does not reference any custom func, so it parses without a FuncMap.
	if err := os.WriteFile(path.Join(pages, "index.html"), []byte(`{{define "body"}}plain{{end}}`), 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}

	templates, err := GenerateTemplates(dir)
	if err != nil {
		t.Fatalf("GenerateTemplates: %v", err)
	}

	key := "pages/index.html"
	if _, ok := templates[key]; !ok {
		t.Fatalf("expected template map to contain key %q, got keys %v", key, keysOf(templates))
	}
}

func keysOf(m map[string]*template.Template) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
