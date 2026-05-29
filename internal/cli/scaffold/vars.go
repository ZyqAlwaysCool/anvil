package scaffold

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/mod/module"
)

// validProjectName 约束项目名可直接用作目录名、Go module 与 import path 前缀。
// 第一版选择显式拒绝非法名，而不是自动 slugify，避免生成与用户预期不一致的项目。
var validProjectName = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

const projectNameFormatHint = "project name must contain only lowercase letters, digits, and hyphens; it cannot start or end with a hyphen"

// Vars 是 CLI 模板渲染使用的变量集合。
// ProjectName 决定生成目录名；ModuleName 默认与项目名相同，也可通过 --module 单独指定。
type Vars struct {
	ProjectName string
	ModuleName  string
}

// NewVars 根据项目名与可选 module 名构造模板变量。
// moduleName 为空时，ModuleName 与 ProjectName 相同。
func NewVars(projectName, moduleName string) (Vars, error) {
	name := strings.TrimSpace(projectName)
	if name == "" {
		return Vars{}, fmt.Errorf("project name is required")
	}
	if strings.ContainsAny(name, `/\`) {
		return Vars{}, fmt.Errorf("%s; path separators are not allowed", projectNameFormatHint)
	}
	if name == "." || name == ".." {
		return Vars{}, fmt.Errorf("%s", projectNameFormatHint)
	}
	if !validProjectName.MatchString(name) {
		return Vars{}, fmt.Errorf("%s", projectNameFormatHint)
	}

	mod := strings.TrimSpace(moduleName)
	if mod == "" {
		mod = name
	} else if err := module.CheckPath(mod); err != nil {
		return Vars{}, fmt.Errorf("invalid module name %q: %v", mod, err)
	}

	return Vars{
		ProjectName: name,
		ModuleName:  mod,
	}, nil
}
