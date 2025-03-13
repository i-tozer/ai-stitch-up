package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	ContentExtraction ContentExtractionConfig `json:"content_extraction"`
	SceneGeneration   SceneGenerationConfig   `json:"scene_generation"`
	ImageCreation     ImageCreationConfig     `json:"image_creation"`
	VideoConversion   VideoConversionConfig   `json:"video_conversion"`
	LyricCreation     LyricCreationConfig     `json:"lyric_creation"`
	MusicGeneration   MusicGenerationConfig   `json:"music_generation"`
	Assembly          AssemblyConfig          `json:"assembly"`
	OutputDir         string                  `json:"output_dir"`
}

// ContentExtractionConfig holds configuration for content extraction
type ContentExtractionConfig struct {
	Source       string `json:"source"`
	ClaudeAPIKey string `json:"claude_api_key"`
}

// SceneGenerationConfig holds configuration for scene generation
type SceneGenerationConfig struct {
	ClaudeKey string `json:"claude_key"`
	MaxScenes int    `json:"max_scenes"`
}

// ImageCreationConfig holds configuration for image creation
type ImageCreationConfig struct {
	HuggingFaceAPIKey string `json:"huggingface_api_key"`
	HuggingFaceModel  string `json:"huggingface_model"`
	OutputDir         string `json:"output_dir"`
}

// VideoConversionConfig holds configuration for video conversion
type VideoConversionConfig struct {
	RunwayAPIKey          string `json:"runway_api_key"`
	OutputDir             string `json:"output_dir"`
	VideoLength           int    `json:"video_length"` // in seconds
	UseNodeImplementation bool   `json:"use_node_implementation"`
}

// LyricCreationConfig holds configuration for lyric creation
type LyricCreationConfig struct {
	ClaudeKey string `json:"claude_key"`
}

// MusicGenerationConfig holds configuration for music generation
type MusicGenerationConfig struct {
	SunoAPIKey string `json:"suno_api_key"`
	OutputDir  string `json:"output_dir"`
}

// AssemblyConfig holds configuration for final assembly
type AssemblyConfig struct {
	FFMPEGPath string `json:"ffmpeg_path"`
	OutputDir  string `json:"output_dir"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	homeDir, _ := os.UserHomeDir()
	outputDir := filepath.Join(homeDir, "stitch-up-output")

	return Config{
		ContentExtraction: ContentExtractionConfig{
			Source: "https://www.wsj.com",
		},
		SceneGeneration: SceneGenerationConfig{
			MaxScenes: 5,
		},
		ImageCreation: ImageCreationConfig{
			OutputDir: filepath.Join(outputDir, "images"),
		},
		VideoConversion: VideoConversionConfig{
			OutputDir:   filepath.Join(outputDir, "videos"),
			VideoLength: 10,
		},
		MusicGeneration: MusicGenerationConfig{
			OutputDir: filepath.Join(outputDir, "music"),
		},
		Assembly: AssemblyConfig{
			FFMPEGPath: "ffmpeg",
			OutputDir:  filepath.Join(outputDir, "final"),
		},
		OutputDir: outputDir,
	}
}

// Load loads the configuration from file or environment
func Load() (Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	config := DefaultConfig()

	// Try to load from config file
	configPath := os.Getenv("STITCH_UP_CONFIG")
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			configPath = filepath.Join(homeDir, ".stitch-up.json")
		}
	}

	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			file, err := os.Open(configPath)
			if err == nil {
				defer file.Close()
				decoder := json.NewDecoder(file)
				if err := decoder.Decode(&config); err != nil {
					return config, fmt.Errorf("error decoding config file: %w", err)
				}
			}
		}
	}

	// Override with environment variables if present
	if source := os.Getenv("BBC_URL"); source != "" {
		config.ContentExtraction.Source = source
	}

	if apiKey := os.Getenv("CLAUDE_API_KEY"); apiKey != "" {
		config.ContentExtraction.ClaudeAPIKey = apiKey
		config.SceneGeneration.ClaudeKey = apiKey
		config.LyricCreation.ClaudeKey = apiKey
	}

	if apiKey := os.Getenv("HUGGINGFACE_API_KEY"); apiKey != "" {
		config.ImageCreation.HuggingFaceAPIKey = apiKey
	}

	if model := os.Getenv("HUGGINGFACE_MODEL"); model != "" {
		config.ImageCreation.HuggingFaceModel = model
	} else {
		// Default to a popular stable diffusion model if not specified
		config.ImageCreation.HuggingFaceModel = "stabilityai/stable-diffusion-xl-base-1.0"
	}

	if apiKey := os.Getenv("RUNWAY_API_KEY"); apiKey != "" {
		config.VideoConversion.RunwayAPIKey = apiKey
	}

	if apiKey := os.Getenv("SUNO_API_KEY"); apiKey != "" {
		config.MusicGeneration.SunoAPIKey = apiKey
	}

	if outputDir := os.Getenv("OUTPUT_DIR"); outputDir != "" {
		config.OutputDir = outputDir
		config.ImageCreation.OutputDir = filepath.Join(outputDir, "images")
		config.VideoConversion.OutputDir = filepath.Join(outputDir, "videos")
		config.MusicGeneration.OutputDir = filepath.Join(outputDir, "music")
		config.Assembly.OutputDir = filepath.Join(outputDir, "final")
	}

	// Create output directories
	os.MkdirAll(config.OutputDir, 0755)
	os.MkdirAll(config.ImageCreation.OutputDir, 0755)
	os.MkdirAll(config.VideoConversion.OutputDir, 0755)
	os.MkdirAll(config.MusicGeneration.OutputDir, 0755)
	os.MkdirAll(config.Assembly.OutputDir, 0755)

	return config, nil
}

// LoadForTest loads the configuration for testing, ensuring .env is loaded
func LoadForTest() (Config, error) {
	// Try to load from project root and test directory
	godotenv.Load()
	godotenv.Load("../../.env")

	return Load()
}
