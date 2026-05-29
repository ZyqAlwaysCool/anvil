//go:build ignore

package prompt

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	sectionSystem = "+++system"
	sectionUser   = "+++user"
)

// Template 固定分为 system/user 两段，避免业务写出不一致协议。
type Template struct {
	System string
	User   string
}

type RenderedPrompt struct {
	System string
	User   string
}

// Load 读取固定 +++system/+++user 协议模板；缺段或缺变量在 Render 阶段直接失败。
func Load(dir, name string) (*Template, error) {
	path := filepath.Join(dir, name+".md")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read prompt template: %w", err)
	}
	return Parse(string(content))
}

func Parse(content string) (*Template, error) {
	sections, err := splitSections(content)
	if err != nil {
		return nil, err
	}
	system, ok := sections[sectionSystem]
	if !ok || strings.TrimSpace(system) == "" {
		return nil, fmt.Errorf("prompt template missing %s section", sectionSystem)
	}
	user, ok := sections[sectionUser]
	if !ok || strings.TrimSpace(user) == "" {
		return nil, fmt.Errorf("prompt template missing %s section", sectionUser)
	}
	return &Template{System: strings.TrimSpace(system), User: strings.TrimSpace(user)}, nil
}

// Render 开启严格校验，缺失变量直接报错，不允许静默回退。
func (t *Template) Render(data any) (*RenderedPrompt, error) {
	system, err := renderSection(t.System, data)
	if err != nil {
		return nil, fmt.Errorf("render system prompt: %w", err)
	}
	user, err := renderSection(t.User, data)
	if err != nil {
		return nil, fmt.Errorf("render user prompt: %w", err)
	}
	return &RenderedPrompt{System: system, User: user}, nil
}

func renderSection(content string, data any) (string, error) {
	tmpl, err := template.New("prompt").Option("missingkey=error").Parse(content)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}

func splitSections(content string) (map[string]string, error) {
	lines := strings.Split(content, "\n")
	sections := make(map[string]string)
	var current string
	var builder strings.Builder

	flush := func() {
		if current == "" {
			return
		}
		sections[current] = builder.String()
		builder.Reset()
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == sectionSystem || trimmed == sectionUser {
			flush()
			current = trimmed
			continue
		}
		if current != "" {
			builder.WriteString(line)
			builder.WriteString("\n")
		}
	}
	flush()
	return sections, nil
}
