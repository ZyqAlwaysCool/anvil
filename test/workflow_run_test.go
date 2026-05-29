package test

import (
	"context"
	"testing"

	"github.com/zyq/anvil/internal/platform/llm/workflow"
)

type echoNode struct {
	key    string
	action string
}

func (n *echoNode) Prep(_ context.Context, state *workflow.State) (any, error) {
	value, _ := state.Get(n.key)
	return value, nil
}

func (n *echoNode) Exec(_ context.Context, prepResult any) (any, error) {
	return prepResult, nil
}

func (n *echoNode) Post(_ context.Context, state *workflow.State, _, execResult any) (string, error) {
	state.Set(n.key+"_done", execResult)
	return n.action, nil
}

func TestWorkflowLinearRun(t *testing.T) {
	state := workflow.NewState()
	state.Set("step", "value")
	flow := workflow.NewFlow("start").
		AddNode("start", &echoNode{key: "step", action: ""})

	if err := flow.Run(context.Background(), state); err != nil {
		t.Fatalf("run workflow: %v", err)
	}
	done, ok := state.Get("step_done")
	if !ok || done != "value" {
		t.Fatalf("unexpected state: %v %v", done, ok)
	}
}

func TestWorkflowBranchRun(t *testing.T) {
	state := workflow.NewState()
	state.Set("branch", true)
	yesNode := &echoNode{key: "branch", action: "yes"}
	noNode := &echoNode{key: "branch", action: ""}
	flow := workflow.NewFlow("start").
		AddNode("start", yesNode).
		AddNode("yes", noNode).
		AddEdge("start", "yes", "yes")

	if err := flow.Run(context.Background(), state); err != nil {
		t.Fatalf("run workflow: %v", err)
	}
}

func TestWorkflowUndefinedAction(t *testing.T) {
	state := workflow.NewState()
	flow := workflow.NewFlow("start").
		AddNode("start", &echoNode{key: "step", action: "missing"})

	err := flow.Run(context.Background(), state)
	if err == nil {
		t.Fatal("expected undefined action error")
	}
}
