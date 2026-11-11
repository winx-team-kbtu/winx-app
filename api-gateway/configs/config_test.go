package configs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// loadConfig writes a config.dev.yaml with the given pilot block into a temp
// directory, runs InitConfig from there and returns the resulting pilot mode.
func loadConfig(t *testing.T, pilotBlock string) bool {
	t.Helper()

	dir := t.TempDir()
	yaml := "app:\n  name: api-gateway\n  url: 0.0.0.0:3000\n" + pilotBlock
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.dev.yaml"), []byte(yaml), 0o600))

	prev, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))

	t.Cleanup(func() {
		_ = os.Chdir(prev)

		viper.Reset()
	})

	viper.Reset()
	t.Setenv("APP_ENV", "")

	InitConfig()

	return Config.Pilot.Mode
}

func TestPilotMode_DefaultsToEnabledWhenNotConfigured(t *testing.T) {
	t.Setenv("PILOT_MODE", "")
	require.NoError(t, os.Unsetenv("PILOT_MODE"))

	assert.True(t, loadConfig(t, ""), "a missing setting must fail closed and keep Social disabled")
}

func TestPilotMode_ExplicitFalseReEnablesSocial(t *testing.T) {
	require.NoError(t, os.Unsetenv("PILOT_MODE"))

	assert.False(t, loadConfig(t, "pilot:\n  mode: false\n"))
}

func TestPilotMode_EnvOverridesFile(t *testing.T) {
	t.Setenv("PILOT_MODE", "false")
	assert.False(t, loadConfig(t, "pilot:\n  mode: true\n"))

	t.Setenv("PILOT_MODE", "true")
	assert.True(t, loadConfig(t, "pilot:\n  mode: false\n"))
}

func TestDevConfigShipsWithPilotModeEnabled(t *testing.T) {
	require.NoError(t, os.Unsetenv("PILOT_MODE"))

	data, err := os.ReadFile("../config.dev.yaml")
	require.NoError(t, err)

	assert.Contains(t, string(data), "mode: true")
}
