package workflow

import (
	"context"
	"fmt"
)

type Flow struct {
	nodes map[string]Node
	edges map[string]map[string]string
	start string
}

func NewFlow(start string) *Flow {
	return &Flow{
		nodes: make(map[string]Node),
		edges: make(map[string]map[string]string),
		start: start,
	}
}

func (f *Flow) AddNode(name string, node Node) *Flow {
	f.nodes[name] = node
	return f
}

func (f *Flow) AddEdge(from, action, to string) *Flow {
	if _, ok := f.edges[from]; !ok {
		f.edges[from] = make(map[string]string)
	}
	f.edges[from][action] = to
	return f
}

// Run 从 start 节点开始执行，直到没有后继节点；未定义 action 直接报错。
func (f *Flow) Run(ctx context.Context, state *State) error {
	current := f.start
	for current != "" {
		node, ok := f.nodes[current]
		if !ok {
			return fmt.Errorf("workflow node not found: %s", current)
		}

		prepResult, err := node.Prep(ctx, state)
		if err != nil {
			return fmt.Errorf("node %s prep failed: %w", current, err)
		}
		execResult, err := node.Exec(ctx, prepResult)
		if err != nil {
			return fmt.Errorf("node %s exec failed: %w", current, err)
		}
		action, err := node.Post(ctx, state, prepResult, execResult)
		if err != nil {
			return fmt.Errorf("node %s post failed: %w", current, err)
		}
		if action == "" {
			break
		}

		nextEdges, ok := f.edges[current]
		if !ok {
			return fmt.Errorf("workflow node %s returned undefined action: %s", current, action)
		}
		next, ok := nextEdges[action]
		if !ok {
			return fmt.Errorf("workflow node %s returned undefined action: %s", current, action)
		}
		current = next
	}
	return nil
}
