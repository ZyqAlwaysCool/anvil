package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZyqAlwaysCool/anvil/templates"
)

const templateRoot = "project"

// MaterializeProject 将 embed 模板写入目标目录。
// templates/project/ 只包含生成后项目真正需要的内容，而不是整仓拷贝再删文件。
func MaterializeProject(dest string, vars Vars, force bool) error {
	if err := ensureTargetDir(dest, force); err != nil {
		return err
	}

	return fs.WalkDir(templates.Project, templateRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(templateRoot, path)
		if err != nil {
			return fmt.Errorf("resolve template path: %w", err)
		}
		if rel == "." {
			return nil
		}

		targetPath := filepath.Join(dest, rel)
		if entry.IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return fmt.Errorf("create directory %s: %w", targetPath, err)
			}
			return nil
		}

		raw, err := templates.Project.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read template %s: %w", path, err)
		}

		rendered, err := renderContent(path, raw, vars)
		if err != nil {
			return err
		}

		outputPath := mapOutputPath(targetPath)
		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return fmt.Errorf("create parent directory for %s: %w", outputPath, err)
		}

		mode := fs.FileMode(0o644)
		if info, statErr := entry.Info(); statErr == nil {
			mode = info.Mode().Perm()
		}

		if err := os.WriteFile(outputPath, rendered, mode); err != nil {
			return fmt.Errorf("write file %s: %w", outputPath, err)
		}
		return nil
	})
}

func mapOutputPath(path string) string {
	switch {
	case strings.HasSuffix(path, ".tmpl"):
		return strings.TrimSuffix(path, ".tmpl")
	default:
		return path
	}
}

func ensureTargetDir(dest string, force bool) error {
	info, err := os.Stat(dest)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dest, 0o755); err != nil {
			return fmt.Errorf("create target directory %s: %w", dest, err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat target directory %s: %w", dest, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("target path %s exists and is not a directory", dest)
	}

	empty, err := dirEmpty(dest)
	if err != nil {
		return err
	}
	if !empty && !force {
		return fmt.Errorf("target directory %s already exists and is not empty; use --force to overwrite", dest)
	}
	return nil
}

func dirEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("read target directory %s: %w", path, err)
	}
	return len(entries) == 0, nil
}
