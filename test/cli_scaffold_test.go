package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ZyqAlwaysCool/anvil/internal/cli/scaffold"
)

func TestNewVarsValidNames(t *testing.T) {
	valid := []string{
		"my-app",
		"demo",
		"agent-platform2",
		"a",
		"abc123",
	}
	for _, name := range valid {
		t.Run(name, func(t *testing.T) {
			vars, err := scaffold.NewVars(name, "")
			if err != nil {
				t.Fatalf("NewVars(%q): %v", name, err)
			}
			if vars.ProjectName != name || vars.ModuleName != name {
				t.Fatalf("vars = %+v, want name %q", vars, name)
			}
		})
	}
}

func TestNewVarsInvalidNames(t *testing.T) {
	invalid := []string{
		"bad name",
		"my_app",
		".demo",
		"demo.",
		"demo/app",
		"DemoApp",
		"中文项目",
		"-demo",
		"demo-",
		"",
		"  ",
	}
	for _, name := range invalid {
		t.Run(name, func(t *testing.T) {
			_, err := scaffold.NewVars(name, "")
			if err == nil {
				t.Fatalf("NewVars(%q) expected error", name)
			}
		})
	}
}

func TestNewVarsCustomModule(t *testing.T) {
	vars, err := scaffold.NewVars("my-app", "github.com/acme/my-app")
	if err != nil {
		t.Fatalf("NewVars: %v", err)
	}
	if vars.ProjectName != "my-app" || vars.ModuleName != "github.com/acme/my-app" {
		t.Fatalf("vars = %+v", vars)
	}
}

func TestNewVarsInvalidModule(t *testing.T) {
	_, err := scaffold.NewVars("my-app", "bad module")
	if err == nil {
		t.Fatal("expected error for invalid module name")
	}
	if !strings.Contains(err.Error(), "invalid module name") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestScaffoldNewWithCustomModule(t *testing.T) {
	outDir := t.TempDir()
	target, err := scaffold.New(scaffold.NewOptions{
		ProjectName: "my-app",
		ModuleName:  "github.com/acme/my-app",
		OutDir:      outDir,
	})
	if err != nil {
		t.Fatalf("scaffold new: %v", err)
	}

	mod, err := os.ReadFile(filepath.Join(target, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !strings.HasPrefix(string(mod), "module github.com/acme/my-app\n") {
		t.Fatalf("unexpected go.mod:\n%s", mod)
	}

	server, err := os.ReadFile(filepath.Join(target, "cmd/server/main.go"))
	if err != nil {
		t.Fatalf("read server main: %v", err)
	}
	if !strings.Contains(string(server), `"github.com/acme/my-app/internal/platform/app"`) {
		t.Fatalf("server import not rendered: %s", server)
	}
}

func TestScaffoldNewRejectsSpaceInProjectName(t *testing.T) {
	_, err := scaffold.New(scaffold.NewOptions{
		ProjectName: "bad name",
		OutDir:      t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected error for project name with space")
	}
	if !strings.Contains(err.Error(), "lowercase letters, digits, and hyphens") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestScaffoldNewCreatesProject(t *testing.T) {
	outDir := t.TempDir()
	target, err := scaffold.New(scaffold.NewOptions{
		ProjectName: "demo-app",
		OutDir:      outDir,
	})
	if err != nil {
		t.Fatalf("scaffold new: %v", err)
	}

	want := filepath.Join(outDir, "demo-app")
	if target != want {
		t.Fatalf("target = %q, want %q", target, want)
	}

	assertPathExists(t, filepath.Join(target, "cmd/server/main.go"))
	assertPathExists(t, filepath.Join(target, "cmd/worker/main.go"))
	assertPathExists(t, filepath.Join(target, "go.mod"))
	assertPathNotExists(t, filepath.Join(target, "cmd/anvil"))
	assertPathNotExists(t, filepath.Join(target, "internal/cli"))
	assertPathNotExists(t, filepath.Join(target, "templates"))
	assertPathNotExists(t, filepath.Join(target, "codex-docs"))

	mod, err := os.ReadFile(filepath.Join(target, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !strings.HasPrefix(string(mod), "module demo-app\n") {
		t.Fatalf("unexpected go.mod:\n%s", mod)
	}

	server, err := os.ReadFile(filepath.Join(target, "cmd/server/main.go"))
	if err != nil {
		t.Fatalf("read server main: %v", err)
	}
	if !strings.Contains(string(server), `"demo-app/internal/platform/app"`) {
		t.Fatalf("server import not rendered: %s", server)
	}

	env, err := os.ReadFile(filepath.Join(target, "configs/.env.example"))
	if err != nil {
		t.Fatalf("read env example: %v", err)
	}
	if !strings.Contains(string(env), "APP_NAME=demo-app") {
		t.Fatalf("env example not rendered: %s", env)
	}

	makefile, err := os.ReadFile(filepath.Join(target, "Makefile"))
	if err != nil {
		t.Fatalf("read Makefile: %v", err)
	}
	if strings.Contains(string(makefile), "run: build") {
		t.Fatalf("make run should not depend on full build target:\n%s", makefile)
	}
}

func TestScaffoldNewRejectsExistingNonEmptyDir(t *testing.T) {
	outDir := t.TempDir()
	target := filepath.Join(outDir, "demo-app")
	if err := os.MkdirAll(filepath.Join(target, "existing"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	_, err := scaffold.New(scaffold.NewOptions{
		ProjectName: "demo-app",
		OutDir:      outDir,
	})
	if err == nil {
		t.Fatal("expected error for non-empty directory")
	}
}

func TestScaffoldNewForceAllowsExistingDir(t *testing.T) {
	outDir := t.TempDir()
	target := filepath.Join(outDir, "demo-app")
	if err := os.MkdirAll(filepath.Join(target, "existing"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	_, err := scaffold.New(scaffold.NewOptions{
		ProjectName: "demo-app",
		OutDir:      outDir,
		Force:       true,
	})
	if err != nil {
		t.Fatalf("scaffold new with force: %v", err)
	}
}

func TestScaffoldNewRejectsEmptyName(t *testing.T) {
	_, err := scaffold.New(scaffold.NewOptions{ProjectName: "  "})
	if err == nil {
		t.Fatal("expected error for empty project name")
	}
}

func assertPathExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected path %s to exist: %v", path, err)
	}
}

func assertPathNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected path %s to not exist", path)
	}
}
