package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_Defaults(t *testing.T) {
	os.Unsetenv("HOST")
	os.Unsetenv("PORT")
	os.Unsetenv("DEV_MODE")

	cfg := LoadConfig()

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 8000, cfg.Port)
	assert.False(t, cfg.DevMode)
}

func TestLoadConfig_FromEnv(t *testing.T) {
	os.Setenv("HOST", "0.0.0.0")
	os.Setenv("PORT", "9090")
	os.Setenv("DEV_MODE", "true")
	t.Cleanup(func() {
		os.Unsetenv("HOST")
		os.Unsetenv("PORT")
		os.Unsetenv("DEV_MODE")
	})

	cfg := LoadConfig()

	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, 9090, cfg.Port)
	assert.True(t, cfg.DevMode)
}

func TestLoadConfig_InvalidPort(t *testing.T) {
	os.Setenv("PORT", "not-a-number")
	t.Cleanup(func() { os.Unsetenv("PORT") })

	cfg := LoadConfig()

	assert.Equal(t, 8000, cfg.Port)
}
