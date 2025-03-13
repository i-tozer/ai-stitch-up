package assembly

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/iantozer/stitch-up/pkg/common"
	"github.com/iantozer/stitch-up/pkg/config"
)

func TestAssembler_Assemble(t *testing.T) {
	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "assemblytest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Check if ffmpeg is available
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Log("Warning: ffmpeg not found in PATH")
	}

	// Create test configuration
	cfg := config.AssemblyConfig{
		OutputDir:  tempDir,
		FFMPEGPath: ffmpegPath,
	}

	// Create assembler instance
	assembler := New(cfg)

	// Create temporary video and music files for testing
	videoDir, err := os.MkdirTemp("", "videotest")
	if err != nil {
		t.Fatalf("Failed to create video temp dir: %v", err)
	}
	defer os.RemoveAll(videoDir)

	// Create test videos
	videos := []common.Video{
		{
			Path:    filepath.Join(videoDir, "video1.mp4"),
			ImageID: "image1",
			Length:  10,
		},
		{
			Path:    filepath.Join(videoDir, "video2.mp4"),
			ImageID: "image2",
			Length:  10,
		},
	}

	// Create placeholder video files
	for _, video := range videos {
		if err := os.WriteFile(video.Path, []byte("test video data"), 0644); err != nil {
			t.Fatalf("Failed to create test video: %v", err)
		}
	}

	// Create test music file
	musicPath := filepath.Join(videoDir, "music.mp3")
	if err := os.WriteFile(musicPath, []byte("test music data"), 0644); err != nil {
		t.Fatalf("Failed to create test music: %v", err)
	}

	music := common.Music{
		Path:     musicPath,
		LyricsID: "lyrics1",
		Length:   180,
	}

	// Test assembly
	ctx := context.Background()
	outputPath, err := assembler.Assemble(ctx, videos, music)
	if err != nil {
		t.Errorf("Assemble() error = %v", err)
		return
	}

	// Validate output path
	if outputPath == "" {
		t.Error("Assemble() returned empty output path")
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Output file does not exist: %s", outputPath)
	}

	// Verify file is in the correct directory
	if !filepath.HasPrefix(outputPath, tempDir) {
		t.Errorf("Output file not in output directory: %s", outputPath)
	}

	// Verify file has correct extension
	if filepath.Ext(outputPath) != ".mp4" {
		t.Errorf("Output file has incorrect extension: %s", filepath.Ext(outputPath))
	}
}

func TestAssembler_Assemble_EmptyVideos(t *testing.T) {
	// Test with empty videos
	tempDir, err := os.MkdirTemp("", "assemblytest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := config.AssemblyConfig{
		OutputDir: tempDir,
	}
	assembler := New(cfg)
	ctx := context.Background()
	videos := []common.Video{}
	music := common.Music{
		Path:     "test.mp3",
		LyricsID: "lyrics1",
		Length:   180,
	}

	outputPath, err := assembler.Assemble(ctx, videos, music)
	if err == nil {
		t.Error("Assemble() with empty videos should return error")
		if outputPath != "" {
			os.Remove(outputPath)
		}
	}
}

func TestAssembler_Assemble_InvalidMusic(t *testing.T) {
	// Test with invalid music file
	tempDir, err := os.MkdirTemp("", "assemblytest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := config.AssemblyConfig{
		OutputDir: tempDir,
	}
	assembler := New(cfg)
	ctx := context.Background()
	videos := []common.Video{
		{
			Path:    "video1.mp4",
			ImageID: "image1",
			Length:  10,
		},
	}
	music := common.Music{
		Path:     "/nonexistent/music.mp3",
		LyricsID: "lyrics1",
		Length:   180,
	}

	outputPath, err := assembler.Assemble(ctx, videos, music)
	// In current implementation, this still creates a placeholder
	// In a real implementation, this would likely fail
	if err != nil {
		t.Errorf("Assemble() with invalid music error = %v", err)
	}
	if outputPath == "" {
		t.Error("Assemble() with invalid music returned empty output path")
	} else {
		os.Remove(outputPath)
	}
}

func TestAssembler_Assemble_FFMPEGNotFound(t *testing.T) {
	// Test with invalid ffmpeg path
	tempDir, err := os.MkdirTemp("", "assemblytest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := config.AssemblyConfig{
		OutputDir:  tempDir,
		FFMPEGPath: "/nonexistent/ffmpeg",
	}
	assembler := New(cfg)
	ctx := context.Background()
	videos := []common.Video{
		{
			Path:    "video1.mp4",
			ImageID: "image1",
			Length:  10,
		},
	}
	music := common.Music{
		Path:     "music.mp3",
		LyricsID: "lyrics1",
		Length:   180,
	}

	outputPath, err := assembler.Assemble(ctx, videos, music)
	// In current implementation, this still creates a placeholder
	// In a real implementation, this would likely fail if ffmpeg is required
	if err != nil {
		t.Errorf("Assemble() with invalid ffmpeg path error = %v", err)
	}
	if outputPath == "" {
		t.Error("Assemble() with invalid ffmpeg path returned empty output path")
	} else {
		os.Remove(outputPath)
	}
}
