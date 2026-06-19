package manifest

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	expectedAPIVersion = "agentvoir.dev/v1alpha1"
	expectedKind       = "Agent"
)

// Parse decodes and validates an Agent YAML manifest.
func Parse(data []byte) (*Document, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	var doc Document
	if err := root.Decode(&doc); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	lines := fieldLines(&root)
	if errs := doc.ValidateDetailed(lines); errs.HasIssues() {
		return nil, errs
	}
	return &doc, nil
}

// ValidateDetailed returns structured validation issues for invalid manifests.
func (d *Document) ValidateDetailed(lines map[string]int) ValidationErrors {
	var issues []ValidationIssue

	if d.APIVersion != expectedAPIVersion {
		issues = append(issues, issue("apiVersion", fmt.Sprintf("must be %s", expectedAPIVersion), lines["apiVersion"]))
	}
	if d.Kind != expectedKind {
		issues = append(issues, issue("kind", fmt.Sprintf("must be %s", expectedKind), lines["kind"]))
	}
	if strings.TrimSpace(d.Metadata.Name) == "" {
		issues = append(issues, issue("metadata.name", "is required", lines["metadata.name"]))
	}
	if strings.TrimSpace(d.Metadata.Version) == "" {
		issues = append(issues, issue("metadata.version", "is required", lines["metadata.version"]))
	}
	if strings.TrimSpace(d.Spec.OwnerTeam) == "" {
		issues = append(issues, issue("spec.ownerTeam", "is required", lines["spec.ownerTeam"]))
	}
	if d.Spec.Cache.Mode != "" {
		switch d.Spec.Cache.Mode {
		case "off", "exact_only", "write_only":
		default:
			issues = append(issues, issue("spec.cache.mode", "must be off, exact_only, or write_only", lines["spec.cache.mode"]))
		}
	}
	return ValidationErrors{Issues: issues}
}

func fieldLines(root *yaml.Node) map[string]int {
	lines := make(map[string]int)
	if root == nil || len(root.Content) == 0 {
		return lines
	}
	docNode := root.Content[0]
	walkMapping(docNode, "", lines)
	return lines
}

func walkMapping(node *yaml.Node, prefix string, lines map[string]int) {
	if node == nil || node.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		path := keyNode.Value
		if prefix != "" {
			path = prefix + "." + keyNode.Value
		}
		lines[path] = keyNode.Line
		if valueNode.Kind == yaml.MappingNode {
			walkMapping(valueNode, path, lines)
		}
	}
}
