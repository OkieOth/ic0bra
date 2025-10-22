package ic0bra_test

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/okieoth/ic0bra" // adjust import path to match your module name
)

func TestNewFileHistoryProvider_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir) // forces os.UserConfigDir() to use tmp

	provider, err := ic0bra.NewFileHistoryProvider("testApp")

	require.NoError(t, err)
	assert.NotNil(t, provider)

	expectedDir := filepath.Join(tmpDir, "testApp", "history")
	assert.Equal(t, expectedDir, provider.HistDir())

	info, err := os.Stat(expectedDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir(), "expected history directory to exist")
}

func TestGetHistFileName_ReturnsCorrectPath(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir) // forces os.UserConfigDir() to use tmp
	p, err := ic0bra.NewFileHistoryProvider("testApp2")
	require.Nil(t, err)

	filePath := p.GetHistFileName("flag1")
	assert.Equal(t, filepath.Join(tmpDir, "testApp2", "history", "flag1.hist"), filePath)
}

func TestSaveHistAndHasHist(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir) // forces os.UserConfigDir() to use tmp
	p, err := ic0bra.NewFileHistoryProvider("testApp3")
	require.NoError(t, err)
	flagName := "example"

	// Save a new history entry
	err = p.SaveHist(flagName, "first-value")
	require.NoError(t, err)

	assert.True(t, p.HasHist(flagName))
	histPath := p.GetHistFileName("example")
	// Read file back and ensure content matches
	data, err := os.ReadFile(histPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "first-value")

	// Save the same value again (should not duplicate)
	err = p.SaveHist(flagName, "first-value")
	require.NoError(t, err)

	// Count lines, should be exactly 1
	file, err := os.Open(histPath)
	require.NoError(t, err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}
	assert.Equal(t, 1, lineCount)
}

func TestGetHistContent_ReadsLines(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir) // forces os.UserConfigDir() to use tmp
	p, err := ic0bra.NewFileHistoryProvider("testApp3")
	require.NoError(t, err)
	flagName := "color"

	histPath := p.GetHistFileName(flagName)
	content := []string{"red", "green", "blue"}

	err = os.WriteFile(histPath, []byte(fmt.Sprintf("%s\n%s\n%s\n", content[0], content[1], content[2])), 0600)
	require.NoError(t, err)

	lines, err := p.GetHistContent(flagName)
	require.NoError(t, err)
	assert.Equal(t, content, lines)
}

func TestGetHistContent_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir) // forces os.UserConfigDir() to use tmp
	p, err := ic0bra.NewFileHistoryProvider("testApp3")
	require.NoError(t, err)

	lines, err := p.GetHistContent("missing")
	assert.Error(t, err)
	assert.Empty(t, lines)
}
