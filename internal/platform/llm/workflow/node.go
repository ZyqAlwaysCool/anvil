package workflow

import "context"

// Node 节点只负责局部语义；分支条件必须由 Post 显式返回 action。
type Node interface {
	Prep(ctx context.Context, state *State) (any, error)
	Exec(ctx context.Context, prepResult any) (any, error)
	Post(ctx context.Context, state *State, prepResult, execResult any) (string, error)
}
