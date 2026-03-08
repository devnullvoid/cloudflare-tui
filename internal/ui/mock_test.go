package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudflare/cloudflare-go"
)

func setupMockCloudflare(t *testing.T) (*httptest.Server, *cloudflare.API) {
	mux := http.NewServeMux()

	// Mock Zones Endpoint
	mux.HandleFunc("/zones", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			resp := map[string]interface{}{
				"success": true,
				"result": []cloudflare.Zone{
					{ID: "123", Name: "example.com", Status: "active"},
					{ID: "456", Name: "pending.io", Status: "pending"},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		if r.Method == http.MethodPost {
			resp := map[string]interface{}{
				"success": true,
				"result":  cloudflare.Zone{ID: "789", Name: "newzone.com", Status: "pending"},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
	})

	// Mock Single Zone Operations
	mux.HandleFunc("/zones/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			resp := map[string]interface{}{"success": true}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		if r.Method == http.MethodPut && r.URL.Path == "/zones/456/activation_check" {
			resp := map[string]interface{}{
				"success": true,
				"result":  cloudflare.Zone{ID: "456", Name: "pending.io", Status: "pending"},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
	})

	// Mock DNS Records Endpoint
	mux.HandleFunc("/zones/123/dns_records", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"success": true,
			"result": []cloudflare.DNSRecord{
				{ID: "rec1", Name: "test.example.com", Type: "A", Content: "1.2.3.4"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(mux)

	// Initialize client pointing to the mock server
	api, err := cloudflare.New("deadbeef", "test@example.com", cloudflare.BaseURL(server.URL))
	if err != nil {
		t.Fatalf("failed to create cloudflare client: %v", err)
	}

	return server, api
}
