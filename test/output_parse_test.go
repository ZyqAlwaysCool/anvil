package test

import (
	"strings"
	"testing"

	"github.com/ZyqAlwaysCool/anvil/internal/platform/llm/output"
)

type sampleOutput struct {
	Greeting string `json:"greeting"`
}

func TestOutputParseSuccess(t *testing.T) {
	value, err := output.Parse[sampleOutput](`{"greeting":"hello"}`)
	if err != nil {
		t.Fatalf("parse output: %v", err)
	}
	if value.Greeting != "hello" {
		t.Fatalf("unexpected value: %+v", value)
	}
}

func TestOutputParseNonJSON(t *testing.T) {
	_, err := output.Parse[sampleOutput]("not-json")
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "summary=") {
		t.Fatalf("expected summary in error: %v", err)
	}
}

func TestOutputSchemaSuccess(t *testing.T) {
	schema, err := output.Schema[sampleOutput]()
	if err != nil {
		t.Fatalf("generate schema: %v", err)
	}
	if schema["type"] == nil {
		t.Fatal("schema type should exist")
	}
}
