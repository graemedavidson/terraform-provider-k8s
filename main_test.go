package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProcessContentNameOnly(t *testing.T) {
	result, err := processContent(`{"metadata": {"name": "foo"}}`, "my_name", "my_namespace", "my_kind")
	if err != nil {
		t.Fatalf("Error processing content: %v", err)
	}

	assert.Equal(t, "kind: my_kind\nmetadata:\n  name: my_name\n  namespace: my_namespace\n", result)
}

func TestProcessContentAdditional(t *testing.T) {
	result, err := processContent(`{"metadata": {"name": "foo", "test": "sdf"}}`, "my_name", "my_namespace", "my_kind")
	if err != nil {
		t.Fatalf("Error processing content: %v", err)
	}

	assert.Equal(t, "kind: my_kind\nmetadata:\n  name: my_name\n  namespace: my_namespace\n  test: sdf\n", result)
}

func TestProcessContentAdditionalNested(t *testing.T) {
	result, err := processContent(`{"metadata": {"name": "foo", "test": {"foo": "bar"}}}`, "my_name", "my_namespace", "my_kind")
	if err != nil {
		t.Fatalf("Error processing content: %v", err)
	}

	assert.Equal(t, "kind: my_kind\nmetadata:\n  name: my_name\n  namespace: my_namespace\n  test:\n    foo: bar\n", result)
}

func TestProcessContentEmpty(t *testing.T) {
	result, err := processContent(`{}`, "my_name", "my_namespace", "my_kind")
	if err != nil {
		t.Fatalf("Error processing content: %v", err)
	}

	assert.Equal(t, "kind: my_kind\nmetadata:\n  name: my_name\n  namespace: my_namespace\n", result)
}

func TestProcessIllegal(t *testing.T) {
	result, err := processContent(`{"metadata": `, "my_name", "my_namespace", "my_kind")
	assert.NotNil(t, err)
	assert.Equal(t, result, "")
}

func TestProcessContentNoNamespace(t *testing.T) {
	result, err := processContent(`{"metadata": {"name": "foo", "test": "sdf"}}`, "my_name", "", "my_kind")
	if err != nil {
		t.Fatalf("Error processing content: %v", err)
	}

	assert.Equal(t, "kind: my_kind\nmetadata:\n  name: my_name\n  test: sdf\n", result)
}

func TestProcessContentNoNamespaceProvided(t *testing.T) {
	result, err := processContent(`{"metadata": {"namespace": "bar", "name": "foo", "test": "sdf"}}`, "my_name", "", "my_kind")
	assert.NotNil(t, err)
	assert.Equal(t, result, "")
}