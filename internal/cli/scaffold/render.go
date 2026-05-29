package scaffold

import (
	"bytes"
	"strings"
)

const scaffoldBuildIgnorePrefix = "//go:build ignore\n\n"

// stripScaffoldBuildIgnore removes the scaffold-only build tag before writing generated projects.
func stripScaffoldBuildIgnore(raw []byte) []byte {
	if bytes.HasPrefix(raw, []byte(scaffoldBuildIgnorePrefix)) {
		return bytes.TrimPrefix(raw, []byte(scaffoldBuildIgnorePrefix))
	}
	return raw
}

// renderContent 对模板文件做占位变量替换。
// 模板源码使用 anvil-scaffold-template / anvil-scaffold-project 作为稳定占位，
// 以便 templates/project 在脚手架仓库内可被 go:build ignore 的 Go 文件通过编译检查。
func renderContent(_ string, raw []byte, vars Vars) ([]byte, error) {
	content := string(stripScaffoldBuildIgnore(raw))
	content = strings.ReplaceAll(content, "anvil-scaffold-template", vars.ModuleName)
	content = strings.ReplaceAll(content, "anvil-scaffold-project", vars.ProjectName)
	content = strings.ReplaceAll(content, "{{.ModuleName}}", vars.ModuleName)
	content = strings.ReplaceAll(content, "{{.ProjectName}}", vars.ProjectName)
	return []byte(content), nil
}
