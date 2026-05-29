package scaffold

import (
	"fmt"
	"strings"
)

// SplitNewArgs 将项目名与 flag 分离，允许 flag 出现在项目名前后。
// 标准库 flag 在首个非 flag 参数后会停止解析，因此 CLI 入口需先手动拆分。
func SplitNewArgs(args []string) (projectName string, flagArgs []string, showHelp bool, err error) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h", arg == "--help":
			showHelp = true
		case arg == "--out":
			if i+1 >= len(args) {
				return "", nil, false, fmt.Errorf("--out requires a value")
			}
			flagArgs = append(flagArgs, arg, args[i+1])
			i++
		case strings.HasPrefix(arg, "--out="):
			flagArgs = append(flagArgs, arg)
		case arg == "--force":
			flagArgs = append(flagArgs, arg)
		case arg == "--module":
			if i+1 >= len(args) {
				return "", nil, false, fmt.Errorf("--module requires a value")
			}
			flagArgs = append(flagArgs, arg, args[i+1])
			i++
		case strings.HasPrefix(arg, "--module="):
			flagArgs = append(flagArgs, arg)
		case strings.HasPrefix(arg, "-"):
			return "", nil, false, fmt.Errorf("unknown flag %q", arg)
		default:
			if projectName != "" {
				return "", nil, false, fmt.Errorf("multiple project names")
			}
			projectName = arg
		}
	}
	return projectName, flagArgs, showHelp, nil
}
