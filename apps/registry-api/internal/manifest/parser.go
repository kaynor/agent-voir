package manifest

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	expectedAPIVersion = "agentvoir.dev/v1alpha1"
	expectedKind       = "Agent"
)

var ErrInvalidManifest = errors.New("invalid agent manifest")

// Parse decodes and validates an Agent YAML manifest.
func Parse(data []byte) (*Document, error) {
	var doc Document
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	if err := doc.Validate(); err != nil {
		return nil, err
	}
	return &doc, nil
}

// Validate checks required manifest fields and supported apiVersion/kind.
func (d *Document) Validate() error {
	if d.APIVersion != expectedAPIVersion {
		return fmt.Errorf("%w: apiVersion must be %s", ErrInvalidManifest, expectedAPIVersion)
	}
	if d.Kind != expectedKind {
		return fmt.Errorf("%w: kind must be %s", ErrInvalidManifest, expectedKind)
	}
	if strings.TrimSpace(d.Metadata.Name) == "" {
		return fmt.Errorf("%w: metadata.name is required", ErrInvalidManifest)
	}
	if strings.TrimSpace(d.Metadata.Version) == "" {
		return fmt.Errorf("%w: metadata.version is required", ErrInvalidManifest)
	}
	if strings.TrimSpace(d.Spec.OwnerTeam) == "" {
		return fmt.Errorf("%w: spec.ownerTeam is required", ErrInvalidManifest)
	}
	return nil
}
