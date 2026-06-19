package manifest_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/agents"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/budgets"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/dependencies"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/manifest"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/modelroutes"
)

func TestParseExampleManifest(t *testing.T) {
	path := filepath.Join("..", "..", "..", "..", "examples", "agents", "customer-support-agent.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read example manifest: %v", err)
	}

	doc, err := manifest.Parse(data)
	if err != nil {
		t.Fatalf("parse manifest: %v", err)
	}
	if doc.Metadata.Name != "customer-support-agent" {
		t.Fatalf("name = %q", doc.Metadata.Name)
	}
	if doc.Spec.Models.Primary.Model != "gpt-4.1-mini" {
		t.Fatalf("primary model = %q", doc.Spec.Models.Primary.Model)
	}
	if len(doc.Spec.Dependencies.Tools) != 2 {
		t.Fatalf("tools = %d, want 2", len(doc.Spec.Dependencies.Tools))
	}
}

func TestParseValidationErrors(t *testing.T) {
	yaml := []byte(`apiVersion: wrong/v1
kind: Agent
metadata:
  name: ""
  version: ""
spec:
  ownerTeam: ""
`)
	_, err := manifest.Parse(yaml)
	if err == nil {
		t.Fatal("expected validation error")
	}
	var validation manifest.ValidationErrors
	if !errors.As(err, &validation) {
		t.Fatalf("expected ValidationErrors, got %T: %v", err, err)
	}
	if len(validation.Issues) < 3 {
		t.Fatalf("issues = %d, want at least 3", len(validation.Issues))
	}
}

func TestRegisterFromManifest(t *testing.T) {
	yaml := []byte(`apiVersion: agentvoir.dev/v1alpha1
kind: Agent
metadata:
  name: demo-agent
  version: 1.0.0
spec:
  ownerTeam: platform
  environment: staging
  riskLevel: low
  models:
    primary:
      provider: openai
      model: gpt-4.1-mini
  budget:
    monthlyUsd: 100
  dependencies:
    tools:
      - search
`)

	stores := manifest.Stores{
		Agents:       agents.NewMemoryStore(),
		Dependencies: dependencies.NewMemoryStore(),
		Budgets:      budgets.NewMemoryStore(),
		ModelRoutes:  modelroutes.NewMemoryStore(),
	}

	result, err := manifest.RegisterYAML(stores, yaml)
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if result.Agent.AgentID != "demo-agent" {
		t.Fatalf("agent_id = %q", result.Agent.AgentID)
	}
	if len(result.Dependencies) != 1 {
		t.Fatalf("dependencies = %d, want 1", len(result.Dependencies))
	}
	if result.Budget == nil {
		t.Fatal("expected budget")
	}
	if result.ModelRoute == nil {
		t.Fatal("expected model route")
	}
}
