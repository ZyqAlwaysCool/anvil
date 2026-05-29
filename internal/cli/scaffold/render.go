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
// 第一版只做简单字符串模板，不引入复杂模板引擎。
func renderContent(_ string, raw []byte, vars Vars) ([]byte, error) {
	content := string(stripScaffoldBuildIgnore(raw))
	if !strings.Contains(content, "{{.ModuleName}}") && !strings.Contains(content, "{{.ProjectName}}") {
		return []byte(content), nil
	}
	content = strings.ReplaceAll(content, "{{.ModuleName}}", vars.ModuleName)
	content = strings.ReplaceAll(content, "{{.ProjectName}}", vars.ProjectName)
	return []byte(content), nil
}
