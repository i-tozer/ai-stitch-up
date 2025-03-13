package musicgeneration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/iantozer/stitch-up/pkg/common"
	"github.com/iantozer/stitch-up/pkg/config"
)

func TestGenerator_Generate(t *testing.T) {
	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "musictest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test configuration
	cfg := config.MusicGenerationConfig{
		OutputDir: tempDir,
	}

	// Create generator instance
	generator := New(cfg)

	// Create test lyrics
	lyrics := common.Lyrics{
		Title: "News of the Day: January 1, 2024",
		Content: `VERSE 1:
Headlines flash across the screen
Stories of a world unseen

CHORUS:
This is the news of today
Moments that will fade away

BRIDGE:
In a world that's changing fast
Some things are meant to last`,
	}

	// Test music generation
	ctx := context.Background()
	music, err := generator.Generate(ctx, lyrics)
	if err != nil {
		t.Errorf("Generate() error = %v", err)
		return
	}

	// Validate music structure
	if music.Path == "" {
		t.Error("Generate() returned empty path")
	}
	if music.LyricsID == "" {
		t.Error("Generate() returned empty lyrics ID")
	}
	if music.Length <= 0 {
		t.Error("Generate() returned invalid length")
	}

	// Verify file exists
	if _, err := os.Stat(music.Path); os.IsNotExist(err) {
		t.Errorf("Music file does not exist: %s", music.Path)
	}

	// Verify file is in the correct directory
	if !filepath.HasPrefix(music.Path, tempDir) {
		t.Errorf("Music file not in output directory: %s", music.Path)
	}

	// Verify file has correct extension
	if filepath.Ext(music.Path) != ".mp3" {
		t.Errorf("Music file has incorrect extension: %s", filepath.Ext(music.Path))
	}
}

func TestGenerator_Generate_EmptyLyrics(t *testing.T) {
	// Test with empty lyrics
	tempDir, err := os.MkdirTemp("", "musictest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := config.MusicGenerationConfig{
		OutputDir: tempDir,
	}
	generator := New(cfg)
	ctx := context.Background()
	lyrics := common.Lyrics{
		Title:   "",
		Content: "",
	}

	music, err := generator.Generate(ctx, lyrics)
	// The current implementation still generates a placeholder
	// In a real implementation, this might return an error
	if err != nil {
		t.Errorf("Generate() with empty lyrics error = %v", err)
	}
	if music.Path == "" || music.LyricsID == "" {
		t.Error("Generate() with empty lyrics returned invalid music")
	}
}

func TestGenerator_Generate_InvalidOutputDir(t *testing.T) {
	// Test with invalid output directory
	cfg := config.MusicGenerationConfig{
		OutputDir: "/nonexistent/directory",
	}
	generator := New(cfg)
	ctx := context.Background()
	lyrics := common.Lyrics{
		Title:   "Test Song",
		Content: "Test lyrics content",
	}

	music, err := generator.Generate(ctx, lyrics)
	// The current implementation creates the directory if it doesn't exist
	// In a real implementation, this might be handled differently
	if err != nil {
		t.Errorf("Generate() with invalid output dir error = %v", err)
	}
	if music.Path == "" {
		t.Error("Generate() with invalid output dir returned empty path")
	}

	// Clean up created directory
	os.RemoveAll("/nonexistent/directory")
}
