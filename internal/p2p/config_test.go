package p2p

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultDataDir(t *testing.T) {
	// Original environment values to restore after test
	origXdg := os.Getenv("XDG_DATA_HOME")
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")

	defer func() {
		os.Setenv("XDG_DATA_HOME", origXdg)
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	tests := []struct {
		name     string
		env      map[string]string
		expected string
	}{
		{
			name: "XDG_DATA_HOME set",
			env: map[string]string{
				"XDG_DATA_HOME": "/tmp/xdg",
				"HOME":          "/tmp/home",
				"USERPROFILE":   "/tmp/userprofile",
			},
			expected: filepath.Join("/tmp/xdg", "sbc"),
		},
		{
			name: "HOME set, XDG_DATA_HOME empty",
			env: map[string]string{
				"XDG_DATA_HOME": "",
				"HOME":          "/tmp/home",
				"USERPROFILE":   "/tmp/userprofile",
			},
			expected: filepath.Join("/tmp/home", ".local/share/sbc"),
		},
		{
			name: "USERPROFILE set, XDG_DATA_HOME and HOME empty",
			env: map[string]string{
				"XDG_DATA_HOME": "",
				"HOME":          "",
				"USERPROFILE":   `C:\Users\Test`,
			},
			expected: filepath.Join(`C:\Users\Test`, "AppData", "Roaming", "SBC"),
		},
		{
			name: "All empty",
			env: map[string]string{
				"XDG_DATA_HOME": "",
				"HOME":          "",
				"USERPROFILE":   "",
			},
			expected: "sbc_data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env
			os.Setenv("XDG_DATA_HOME", "")
			os.Setenv("HOME", "")
			os.Setenv("USERPROFILE", "")

			// Set test env
			for k, v := range tt.env {
				os.Setenv(k, v)
			}

			got := DefaultDataDir()
			if got != tt.expected {
				t.Errorf("DefaultDataDir() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	origEnv := os.Getenv("SBC_BOOTNODES")
	defer os.Setenv("SBC_BOOTNODES", origEnv)

	t.Run("Default boot nodes", func(t *testing.T) {
		os.Setenv("SBC_BOOTNODES", "")
		cfg := DefaultConfig()
		if len(cfg.BootNode) != len(DefaultBootNodes) {
			t.Errorf("expected %d boot nodes, got %d", len(DefaultBootNodes), len(cfg.BootNode))
		}
	})

	t.Run("Environment override", func(t *testing.T) {
		customNode := "/ip4/1.2.3.4/tcp/1234/p2p/QmTest"
		os.Setenv("SBC_BOOTNODES", customNode)
		cfg := DefaultConfig()
		if len(cfg.BootNode) != 1 || cfg.BootNode[0] != customNode {
			t.Errorf("expected custom boot node %s, got %v", customNode, cfg.BootNode)
		}
	})
}
