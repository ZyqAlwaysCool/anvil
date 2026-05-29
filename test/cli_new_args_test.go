package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zyq/anvil/internal/cli/scaffold"
)

func TestSplitNewArgsProjectBeforeFlags(t *testing.T) {
	name, flags, showHelp, err := scaffold.SplitNewArgs([]string{"my-app", "--out", "/tmp", "--force"})
	if err != nil {
		t.Fatalf("split args: %v", err)
	}
	if showHelp {
		t.Fatal("did not expect help")
	}
	if name != "my-app" {
		t.Fatalf("project name = %q", name)
	}
	if len(flags) != 3 {
		t.Fatalf("flags = %v", flags)
	}
}

func TestSplitNewArgsFlagsBeforeProject(t *testing.T) {
	name, flags, showHelp, err := scaffold.SplitNewArgs([]string{"--out", "/tmp", "my-app"})
	if err != nil {
		t.Fatalf("split args: %v", err)
	}
	if showHelp {
		t.Fatal("did not expect help")
	}
	if name != "my-app" {
		t.Fatalf("project name = %q", name)
	}
	if len(flags) != 2 {
		t.Fatalf("flags = %v", flags)
	}
}

func TestSplitNewArgsHelpFlags(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{name: "short before project", args: []string{"-h"}},
		{name: "long before project", args: []string{"--help"}},
		{name: "short after project", args: []string{"my-app", "-h"}},
		{name: "long after project", args: []string{"my-app", "--help"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, showHelp, err := scaffold.SplitNewArgs(tc.args)
			if err != nil {
				t.Fatalf("split args: %v", err)
			}
			if !showHelp {
				t.Fatal("expected help")
			}
		})
	}
}

func TestSplitNewArgsWithModule(t *testing.T) {
	name, flags, showHelp, err := scaffold.SplitNewArgs([]string{"my-app", "--module", "github.com/acme/my-app"})
	if err != nil {
		t.Fatalf("split args: %v", err)
	}
	if showHelp {
		t.Fatal("did not expect help")
	}
	if name != "my-app" {
		t.Fatalf("project name = %q", name)
	}
	if len(flags) != 2 || flags[0] != "--module" || flags[1] != "github.com/acme/my-app" {
		t.Fatalf("flags = %v", flags)
	}
}

func TestSplitNewArgsRejectsUnknownFlag(t *testing.T) {
	_, _, _, err := scaffold.SplitNewArgs([]string{"--unknown", "x", "my-app"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
}

func TestAnvilNewHelp(t *testing.T) {
	repoRoot := repoRoot(t)
	for _, args := range [][]string{
		{"new", "-h"},
		{"new", "--help"},
	} {
		t.Run(args[1], func(t *testing.T) {
			cmd := exec.Command("go", append([]string{"run", "./cmd/anvil"}, args...)...)
			cmd.Dir = repoRoot
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("run anvil help: %v\n%s", err, out)
			}
			if !containsAll(string(out), "anvil new", "--module", "--out", "--force", "<project-name>") {
				t.Fatalf("unexpected help output:\n%s", out)
			}
		})
	}
}

func TestAnvilNewRejectsInvalidProjectName(t *testing.T) {
	repoRoot := repoRoot(t)
	outDir := t.TempDir()
	cmd := exec.Command("go", "run", "./cmd/anvil", "new", "bad name", "--out", outDir)
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected command failure\n%s", out)
	}
	if !containsAll(string(out), "lowercase letters, digits, and hyphens") {
		t.Fatalf("unexpected error output:\n%s", out)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return filepath.Dir(wd)
}

func containsAll(s string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(s, part) {
			return false
		}
	}
	return true
}
