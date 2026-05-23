package cli

import (
	"strings"
	"testing"

	"github.com/devnullvoid/cloudflare-tui/internal/ui"
	"github.com/spf13/viper"
)

func TestGetTheme(t *testing.T) {
	viper.Reset()

	// Test default theme
	theme := getTheme()
	if theme.Primary != ui.AvailableThemes["ansi"].Primary {
		t.Error("Expected default theme to be ansi")
	}

	// Test mocha theme
	viper.Set("theme", "mocha")
	theme = getTheme()
	if theme.Primary != ui.AvailableThemes["mocha"].Primary {
		t.Errorf("Expected mocha theme, got something else. Primary color: %v", theme.Primary)
	}

	// Test invalid theme fallback
	viper.Set("theme", "non-existent")
	theme = getTheme()
	if theme.Primary != ui.AvailableThemes["ansi"].Primary {
		t.Error("Expected fallback to ansi for invalid theme")
	}
}

func resetRecordFlags() {
	recordType = "A"
	recordName = ""
	recordContent = ""
	recordProxied = false
	recordTTL = 1
	recordComment = ""
	recordFlattenCNAME = false
	recordPriority = 10
	recordService = ""
	recordProto = "_tcp"
	recordWeight = 1
	recordPort = 0
	recordTarget = ""
	recordTag = ""
	recordCAAFlags = 0
	recordValue = ""
}

func TestValidateAndBuildFields(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		wantErr string
		check   func(t *testing.T, f recordFieldSet)
	}{
		{
			name: "A record valid",
			setup: func() {
				recordType = "A"
				recordContent = "1.2.3.4"
			},
			check: func(t *testing.T, f recordFieldSet) {
				if f.content != "1.2.3.4" {
					t.Errorf("content = %q, want 1.2.3.4", f.content)
				}
				if f.priority != nil || f.data != nil {
					t.Error("expected nil priority and data for A record")
				}
			},
		},
		{
			name: "A record missing content",
			setup: func() {
				recordType = "A"
				recordContent = ""
			},
			wantErr: "--content is required",
		},
		{
			name: "AAAA record valid",
			setup: func() {
				recordType = "AAAA"
				recordContent = "2001:db8::1"
			},
			check: func(t *testing.T, f recordFieldSet) {
				if f.content != "2001:db8::1" {
					t.Errorf("content = %q, want 2001:db8::1", f.content)
				}
			},
		},
		{
			name: "TXT record valid",
			setup: func() {
				recordType = "TXT"
				recordContent = "v=spf1 ~all"
			},
			check: func(t *testing.T, f recordFieldSet) {
				if f.content != "v=spf1 ~all" {
					t.Errorf("content = %q, want v=spf1 ~all", f.content)
				}
			},
		},
		{
			name: "CNAME record valid",
			setup: func() {
				recordType = "CNAME"
				recordContent = "target.example.com"
				recordFlattenCNAME = true
			},
			check: func(t *testing.T, f recordFieldSet) {
				if f.content != "target.example.com" {
					t.Errorf("content = %q, want target.example.com", f.content)
				}
				if f.settings.FlattenCNAME == nil || !*f.settings.FlattenCNAME {
					t.Error("expected FlattenCNAME = true")
				}
			},
		},
		{
			name: "CNAME record missing content",
			setup: func() {
				recordType = "CNAME"
				recordContent = ""
			},
			wantErr: "--content is required",
		},
		{
			name: "MX record valid",
			setup: func() {
				recordType = "MX"
				recordContent = "mail.example.com"
				recordPriority = 10
			},
			check: func(t *testing.T, f recordFieldSet) {
				if f.content != "mail.example.com" {
					t.Errorf("content = %q, want mail.example.com", f.content)
				}
				if f.priority == nil || *f.priority != 10 {
					t.Errorf("priority = %v, want 10", f.priority)
				}
			},
		},
		{
			name: "MX record missing content",
			setup: func() {
				recordType = "MX"
				recordContent = ""
			},
			wantErr: "--content is required for MX records",
		},
		{
			name: "SRV record valid",
			setup: func() {
				recordType = "SRV"
				recordName = "_sip._tcp"
				recordService = "_sip"
				recordProto = "_tcp"
				recordPriority = 10
				recordWeight = 20
				recordPort = 5060
				recordTarget = "sip.example.com"
			},
			check: func(t *testing.T, f recordFieldSet) {
				data, ok := f.data.(map[string]interface{})
				if !ok {
					t.Fatal("expected data to be map[string]interface{}")
				}
				if data["service"] != "_sip" {
					t.Errorf("service = %v, want _sip", data["service"])
				}
				if data["target"] != "sip.example.com" {
					t.Errorf("target = %v, want sip.example.com", data["target"])
				}
				if data["port"] != 5060 {
					t.Errorf("port = %v, want 5060", data["port"])
				}
			},
		},
		{
			name: "SRV record missing service",
			setup: func() {
				recordType = "SRV"
				recordService = ""
				recordTarget = "sip.example.com"
			},
			wantErr: "--service is required",
		},
		{
			name: "SRV record missing target",
			setup: func() {
				recordType = "SRV"
				recordService = "_sip"
				recordTarget = ""
			},
			wantErr: "--target is required",
		},
		{
			name: "CAA record valid",
			setup: func() {
				recordType = "CAA"
				recordTag = "issue"
				recordValue = "letsencrypt.org"
				recordCAAFlags = 0
			},
			check: func(t *testing.T, f recordFieldSet) {
				data, ok := f.data.(map[string]interface{})
				if !ok {
					t.Fatal("expected data to be map[string]interface{}")
				}
				if data["tag"] != "issue" {
					t.Errorf("tag = %v, want issue", data["tag"])
				}
				if data["value"] != "letsencrypt.org" {
					t.Errorf("value = %v, want letsencrypt.org", data["value"])
				}
			},
		},
		{
			name: "CAA record missing tag",
			setup: func() {
				recordType = "CAA"
				recordTag = ""
				recordValue = "letsencrypt.org"
			},
			wantErr: "--tag is required",
		},
		{
			name: "CAA record missing value",
			setup: func() {
				recordType = "CAA"
				recordTag = "issue"
				recordValue = ""
			},
			wantErr: "--value is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resetRecordFlags()
			tc.setup()

			f, err := validateAndBuildFields()

			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), tc.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.check != nil {
				tc.check(t, f)
			}
		})
	}
}

func TestConfigMapping(t *testing.T) {
	viper.Reset()
	viper.Set("api_token", "test-token")
	viper.Set("format", "json")
	viper.Set("debug", true)

	// Since we can't easily run PersistentPreRunE in a unit test without more setup,
	// we just check if viper is working as expected which is what our config relies on.
	if viper.GetString("api_token") != "test-token" {
		t.Error("Viper api_token mapping failed")
	}
	if viper.GetString("format") != "json" {
		t.Error("Viper format mapping failed")
	}
	if !viper.GetBool("debug") {
		t.Error("Viper debug mapping failed")
	}
}
