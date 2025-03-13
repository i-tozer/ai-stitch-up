package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	imagecreation "github.com/iantozer/stitch-up/pkg/3_imagecreation"
	"github.com/iantozer/stitch-up/pkg/common"
	"github.com/iantozer/stitch-up/pkg/config"
)

func main() {
	// Parse command-line flags
	scenesPath := flag.String("scenes", "output/scenes.json", "Path to the scenes JSON file")
	outputDir := flag.String("output", "output/images", "Directory to save the generated images")
	modelFlag := flag.String("model", "", "Hugging Face model to use (overrides env var and default)")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Override output directory from command-line flag
	cfg.ImageCreation.OutputDir = *outputDir

	// Override model if specified in command-line flag
	if *modelFlag != "" {
		cfg.ImageCreation.HuggingFaceModel = *modelFlag
	}

	// Log the model being used
	log.Printf("Using Hugging Face model: %s", cfg.ImageCreation.HuggingFaceModel)

	// Create image creator
	creator := imagecreation.New(cfg.ImageCreation)

	// Load scenes from file
	scenes, err := loadScenesFromFile(*scenesPath)
	if err != nil {
		log.Fatalf("Failed to load scenes from file: %v", err)
	}

	log.Printf("Loaded %d scenes from %s", len(scenes), *scenesPath)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Generate images
	startTime := time.Now()
	log.Printf("Starting image generation at %s", startTime.Format("15:04:05"))

	images, err := creator.Create(ctx, scenes)
	if err != nil {
		log.Fatalf("Failed to create images: %v", err)
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	log.Printf("Finished image generation at %s (took %s)", endTime.Format("15:04:05"), duration)
	log.Printf("Generated %d images and saved to %s", len(images), *outputDir)

	// Save image metadata to file
	metadataPath := *outputDir + "/metadata.json"
	if err := saveImagesToFile(images, metadataPath); err != nil {
		log.Fatalf("Failed to save image metadata to file: %v", err)
	}

	log.Printf("Saved image metadata to %s", metadataPath)
}

// loadScenesFromFile loads scenes from a JSON file
func loadScenesFromFile(path string) ([]common.Scene, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var scenes []common.Scene
	if err := json.Unmarshal(data, &scenes); err != nil {
		return nil, err
	}

	return scenes, nil
}

// saveImagesToFile saves image metadata to a JSON file
func saveImagesToFile(images []common.Image, path string) error {
	// Convert to JSON
	data, err := json.MarshalIndent(images, "", "  ")
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	dir := path[:len(path)-len("/metadata.json")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(path, data, 0644)
}
