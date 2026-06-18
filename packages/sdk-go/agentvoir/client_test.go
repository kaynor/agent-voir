package agentvoir

import "testing"

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:8081/", "dev")
	if client.BaseURL != "http://localhost:8081" {
		t.Fatalf("unexpected base URL: %s", client.BaseURL)
	}
}
