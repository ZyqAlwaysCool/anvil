package test

import (
	"strings"
	"testing"

	"github.com/ZyqAlwaysCool/anvil/internal/platform/llm/prompt"
)

func TestPromptTemplateRender(t *testing.T) {
	tmpl, err := prompt.Parse(`+++system
hello {{ .Lang }}

+++user
{{ .Name }}`)
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}
	rendered, err := tmpl.Render(map[string]string{
		"Lang": "zh",
		"Name": "world",
	})
	if err != nil {
		t.Fatalf("render template: %v", err)
	}
	if !strings.Contains(rendered.System, "zh") || !strings.Contains(rendered.User, "world") {
		t.Fatalf("unexpected rendered prompt: %+v", rendered)
	}
}

func TestPromptTemplateMissingSystem(t *testing.T) {
	_, err := prompt.Parse(`+++user
hello`)
	if err == nil {
		t.Fatal("expected missing system error")
	}
}

func TestPromptTemplateMissingVariable(t *testing.T) {
	tmpl, err := prompt.Parse(`+++system
hello

+++user
{{ .Name }}`)
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}
	_, err = tmpl.Render(map[string]string{})
	if err == nil {
		t.Fatal("expected missing variable error")
	}
}
