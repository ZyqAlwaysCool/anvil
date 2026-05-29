package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/zyq/anvil/internal/cli/scaffold"
)

const usage = `anvil - generate a runnable Go Agent project from the anvil runtime template

Usage:
  anvil new <project-name> [--module <go-module-path>] [--out <dir>] [--force]

Arguments:
  <project-name>  Project directory name. Also used as the Go module name unless --module is set.

Flags:
  --module  Go module path for go.mod and imports (e.g. github.com/acme/my-app).
            When omitted, the module name equals <project-name>.
  --out     Parent directory for the generated project (default: current directory).
  --force   Overwrite files when the target directory is not empty.

Examples:
  anvil new my-app
  anvil new my-app --module github.com/acme/my-app
  anvil new my-app --out /tmp/workspace
  anvil new my-app --module github.com/acme/my-app --force
`

func main() {
	if len(os.Args) < 2 {
		exitWithUsage(1)
	}

	switch os.Args[1] {
	case "new":
		if err := runNew(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "-h", "--help", "help":
		exitWithUsage(0)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command %q\n\n", os.Args[1])
		exitWithUsage(1)
	}
}

func runNew(args []string) error {
	projectName, flagArgs, showHelp, err := scaffold.SplitNewArgs(args)
	if err != nil {
		return err
	}
	if showHelp {
		fmt.Fprint(os.Stderr, usage)
		return nil
	}

	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	moduleName := fs.String("module", "", "Go module path; defaults to <project-name>")
	outDir := fs.String("out", "", "parent directory for the generated project")
	force := fs.Bool("force", false, "overwrite files when the target directory is not empty")

	fs.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}

	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if projectName == "" {
		fmt.Fprint(os.Stderr, usage)
		return fmt.Errorf("project name is required")
	}

	target, err := scaffold.New(scaffold.NewOptions{
		ProjectName: projectName,
		ModuleName:  *moduleName,
		OutDir:      *outDir,
		Force:       *force,
	})
	if err != nil {
		return err
	}

	fmt.Printf("created project at %s\n", target)
	fmt.Println("next steps:")
	fmt.Println("  cd", target)
	fmt.Println("  cp configs/.env.example .env")
	fmt.Println("  make run")
	return nil
}

func exitWithUsage(code int) {
	fmt.Fprint(os.Stderr, usage)
	os.Exit(code)
}
