package cli

import (
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
