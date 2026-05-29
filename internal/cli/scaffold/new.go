package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

// NewOptions 描述 anvil new 的路径、module 与覆盖行为。
type NewOptions struct {
	ProjectName string
	ModuleName  string
	OutDir      string
	Force       bool
}

// New 生成一个基于 anvil 运行模板的独立项目目录。
// 目录名始终为 ProjectName；Go module 默认与项目名相同，也可通过 ModuleName 覆盖。
func New(opts NewOptions) (string, error) {
	vars, err := NewVars(opts.ProjectName, opts.ModuleName)
	if err != nil {
		return "", err
	}

	outDir := opts.OutDir
	if outDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolve current directory: %w", err)
		}
		outDir = wd
	}

	outDir, err = filepath.Abs(outDir)
	if err != nil {
		return "", fmt.Errorf("resolve output directory: %w", err)
	}

	targetDir := filepath.Join(outDir, vars.ProjectName)
	if err := MaterializeProject(targetDir, vars, opts.Force); err != nil {
		return "", err
	}
	return targetDir, nil
}
