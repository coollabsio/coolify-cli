package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstance_Validate(t *testing.T) {
	tests := []struct {
		name     string
		instance Instance
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid instance",
			instance: Instance{
				Name:  "test",
				FQDN:  "https://app.coolify.io",
				Token: "test-token",
			},
			wantErr: false,
		},
		{
			name: "valid instance with http",
			instance: Instance{
				Name:  "test",
				FQDN:  "http://localhost:8000",
				Token: "test-token",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			instance: Instance{
				Name:  "",
				FQDN:  "https://app.coolify.io",
				Token: "test-token",
			},
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name: "empty FQDN",
			instance: Instance{
				Name:  "test",
				FQDN:  "",
				Token: "test-token",
			},
			wantErr: true,
			errMsg:  "FQDN cannot be empty",
		},
		{
			name: "invalid FQDN (no protocol)",
			instance: Instance{
				Name:  "test",
				FQDN:  "app.coolify.io",
				Token: "test-token",
			},
			wantErr: true,
			errMsg:  "must start with http",
		},
		{
			name: "empty token",
			instance: Instance{
				Name:  "test",
				FQDN:  "https://app.coolify.io",
				Token: "",
			},
			wantErr: true,
			errMsg:  "token cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.instance.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	cfg := New()

	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.Instances)
	assert.NotEmpty(t, cfg.LastUpdateCheckTime)
}

func TestConfig_AddInstance(t *testing.T) {
	t.Run("add first instance makes it default", func(t *testing.T) {
		cfg := New()

		instance := Instance{
			Name:  "test",
			FQDN:  "https://app.coolify.io",
			Token: "test-token",
		}

		err := cfg.AddInstance(instance)

		require.NoError(t, err)
		assert.Len(t, cfg.Instances, 1)
		assert.True(t, cfg.Instances[0].Default)
	})

	t.Run("add second instance keeps first as default", func(t *testing.T) {
		cfg := New()

		cfg.AddInstance(Instance{
			Name:  "first",
			FQDN:  "https://first.io",
			Token: "token1",
		})

		err := cfg.AddInstance(Instance{
			Name:  "second",
			FQDN:  "https://second.io",
			Token: "token2",
		})

		require.NoError(t, err)
		assert.Len(t, cfg.Instances, 2)
		assert.True(t, cfg.Instances[0].Default)
		assert.False(t, cfg.Instances[1].Default)
	})

	t.Run("add instance with default flag", func(t *testing.T) {
		cfg := New()

		cfg.AddInstance(Instance{
			Name:  "first",
			FQDN:  "https://first.io",
			Token: "token1",
		})

		err := cfg.AddInstance(Instance{
			Name:    "second",
			FQDN:    "https://second.io",
			Token:   "token2",
			Default: true,
		})

		require.NoError(t, err)
		assert.False(t, cfg.Instances[0].Default)
		assert.True(t, cfg.Instances[1].Default)
	})

	t.Run("duplicate name returns error", func(t *testing.T) {
		cfg := New()

		cfg.AddInstance(Instance{
			Name:  "test",
			FQDN:  "https://test.io",
			Token: "token1",
		})

		err := cfg.AddInstance(Instance{
			Name:  "test",
			FQDN:  "https://other.io",
			Token: "token2",
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("invalid instance returns error", func(t *testing.T) {
		cfg := New()

		err := cfg.AddInstance(Instance{
			Name:  "",
			FQDN:  "https://test.io",
			Token: "token",
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid instance")
	})
}

func TestConfig_RemoveInstance(t *testing.T) {
	t.Run("remove existing instance", func(t *testing.T) {
		cfg := New()
		cfg.AddInstance(Instance{
			Name:  "first",
			FQDN:  "https://first.io",
			Token: "token1",
		})
		cfg.AddInstance(Instance{
			Name:  "second",
			FQDN:  "https://second.io",
			Token: "token2",
		})

		err := cfg.RemoveInstance("second")

		require.NoError(t, err)
		assert.Len(t, cfg.Instances, 1)
		assert.Equal(t, "first", cfg.Instances[0].Name)
	})

	t.Run("remove default instance makes first default", func(t *testing.T) {
		cfg := New()
		cfg.AddInstance(Instance{
			Name:  "first",
			FQDN:  "https://first.io",
			Token: "token1",
		})
		cfg.AddInstance(Instance{
			Name:  "second",
			FQDN:  "https://second.io",
			Token: "token2",
		})
		cfg.SetDefault("second")

		err := cfg.RemoveInstance("second")

		require.NoError(t, err)
		assert.True(t, cfg.Instances[0].Default)
	})

	t.Run("remove non-existent instance returns error", func(t *testing.T) {
		cfg := New()

		err := cfg.RemoveInstance("nonexistent")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestConfig_GetInstance(t *testing.T) {
	cfg := New()
	cfg.AddInstance(Instance{
		Name:  "test",
		FQDN:  "https://test.io",
		Token: "test-token",
	})

	t.Run("get existing instance", func(t *testing.T) {
		instance, err := cfg.GetInstance("test")

		require.NoError(t, err)
		assert.Equal(t, "test", instance.Name)
		assert.Equal(t, "https://test.io", instance.FQDN)
	})

	t.Run("get non-existent instance", func(t *testing.T) {
		_, err := cfg.GetInstance("nonexistent")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestConfig_GetDefault(t *testing.T) {
	t.Run("get default instance", func(t *testing.T) {
		cfg := New()
		cfg.AddInstance(Instance{
			Name:    "test",
			FQDN:    "https://test.io",
			Token:   "test-token",
			Default: true,
		})

		instance, err := cfg.GetDefault()

		require.NoError(t, err)
		assert.Equal(t, "test", instance.Name)
	})

	t.Run("no default instance", func(t *testing.T) {
		cfg := New()

		_, err := cfg.GetDefault()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no default instance")
	})
}

func TestConfig_SetDefault(t *testing.T) {
	t.Run("set existing instance as default", func(t *testing.T) {
		cfg := New()
		cfg.AddInstance(Instance{
			Name:  "first",
			FQDN:  "https://first.io",
			Token: "token1",
		})
		cfg.AddInstance(Instance{
			Name:  "second",
			FQDN:  "https://second.io",
			Token: "token2",
		})

		err := cfg.SetDefault("second")

		require.NoError(t, err)
		assert.False(t, cfg.Instances[0].Default)
		assert.True(t, cfg.Instances[1].Default)
	})

	t.Run("set non-existent instance returns error", func(t *testing.T) {
		cfg := New()

		err := cfg.SetDefault("nonexistent")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestConfig_UpdateInstanceToken(t *testing.T) {
	t.Run("update existing instance token", func(t *testing.T) {
		cfg := New()
		cfg.AddInstance(Instance{
			Name:  "test",
			FQDN:  "https://test.io",
			Token: "old-token",
		})

		err := cfg.UpdateInstanceToken("test", "new-token")

		require.NoError(t, err)
		instance, _ := cfg.GetInstance("test")
		assert.Equal(t, "new-token", instance.Token)
	})

	t.Run("update non-existent instance", func(t *testing.T) {
		cfg := New()

		err := cfg.UpdateInstanceToken("nonexistent", "token")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("empty token returns error", func(t *testing.T) {
		cfg := New()
		cfg.AddInstance(Instance{
			Name:  "test",
			FQDN:  "https://test.io",
			Token: "old-token",
		})

		err := cfg.UpdateInstanceToken("test", "")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := New()
		cfg.AddInstance(Instance{
			Name:  "test",
			FQDN:  "https://test.io",
			Token: "token",
		})

		err := cfg.Validate()

		require.NoError(t, err)
	})

	t.Run("empty instances", func(t *testing.T) {
		cfg := New()

		err := cfg.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no instances")
	})

	t.Run("invalid instance", func(t *testing.T) {
		cfg := New()
		// Bypass AddInstance validation
		cfg.Instances = append(cfg.Instances, Instance{
			Name:  "",
			FQDN:  "https://test.io",
			Token: "token",
		})

		err := cfg.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "is invalid")
	})

	t.Run("duplicate names", func(t *testing.T) {
		cfg := New()
		// Bypass AddInstance validation
		cfg.Instances = append(cfg.Instances, Instance{
			Name:  "test",
			FQDN:  "https://test1.io",
			Token: "token1",
		})
		cfg.Instances = append(cfg.Instances, Instance{
			Name:  "test",
			FQDN:  "https://test2.io",
			Token: "token2",
		})

		err := cfg.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate")
	})
}

func TestLoadFromFile(t *testing.T) {
	t.Run("load valid config", func(t *testing.T) {
		// Create temp file
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		validConfig := Config{
			Instances: []Instance{
				{
					Name:    "test",
					FQDN:    "https://test.io",
					Token:   "test-token",
					Default: true,
				},
			},
			LastUpdateCheckTime: "2025-10-14T12:00:00Z",
		}

		data, _ := json.Marshal(validConfig)
		os.WriteFile(configPath, data, 0600)

		// Load config
		cfg, err := LoadFromFile(configPath)

		require.NoError(t, err)
		assert.Len(t, cfg.Instances, 1)
		assert.Equal(t, "test", cfg.Instances[0].Name)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := LoadFromFile("/nonexistent/config.json")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		os.WriteFile(configPath, []byte("invalid json"), 0600)

		_, err := LoadFromFile(configPath)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse")
	})
}

func TestSaveToFile(t *testing.T) {
	t.Run("save valid config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		cfg := New()
		cfg.AddInstance(Instance{
			Name:  "test",
			FQDN:  "https://test.io",
			Token: "test-token",
		})

		err := SaveToFile(configPath, cfg)

		require.NoError(t, err)

		// Verify file was created
		assert.FileExists(t, configPath)

		// Verify content
		data, _ := os.ReadFile(configPath)
		var loaded Config
		json.Unmarshal(data, &loaded)
		assert.Len(t, loaded.Instances, 1)
		assert.Equal(t, "test", loaded.Instances[0].Name)
	})

	t.Run("nil config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		err := SaveToFile(configPath, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})
}

func TestCreateDefault(t *testing.T) {
	// Test that CreateDefault sets up proper instances
	cfg := New()

	// Simulate CreateDefault
	cfg.Instances = append(cfg.Instances, Instance{
		Name:    "cloud",
		FQDN:    "https://app.coolify.io",
		Token:   "",
		Default: true,
	})

	cfg.Instances = append(cfg.Instances, Instance{
		Name:  "localhost",
		FQDN:  "http://localhost:8000",
		Token: "root",
	})

	// Verify instances
	assert.Len(t, cfg.Instances, 2)
	assert.Equal(t, "cloud", cfg.Instances[0].Name)
	assert.Equal(t, "localhost", cfg.Instances[1].Name)
	assert.Equal(t, "root", cfg.Instances[1].Token)
	assert.True(t, cfg.Instances[0].Default)
	assert.False(t, cfg.Instances[1].Default)
}

func TestPath(t *testing.T) {
	path := Path()

	// Should not be empty
	assert.NotEmpty(t, path)

	// Should contain "coolify"
	assert.Contains(t, path, "coolify")

	// Should end with config.json
	assert.Contains(t, path, "config.json")

	// On Unix systems, should contain .config
	// On Windows, should contain AppData or Roaming
	if filepath.Separator == '/' {
		assert.Contains(t, path, ".config")
	} else {
		// Windows path should contain either AppData or backslashes
		assert.True(t,
			filepath.Separator == '\\' &&
			(os.Getenv("APPDATA") != "" || filepath.IsAbs(path)),
			"Windows path should be valid",
		)
	}

	t.Logf("Config path: %s", path)
}
