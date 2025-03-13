package main

import (
	"context"
	"flag"
	"log"
	"time"

	scenegeneration "github.com/iantozer/stitch-up/pkg/2_scenegeneration"
	"github.com/iantozer/stitch-up/pkg/common"
	"github.com/iantozer/stitch-up/pkg/config"
)

func main() {
	// Parse command-line flags
	outputPath := flag.String("output", "output/scenes.json", "Path to save the generated scenes")
	maxScenes := flag.Int("max-scenes", 10, "Maximum number of scenes to generate")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Override max scenes from command-line flag
	cfg.SceneGeneration.MaxScenes = *maxScenes

	// Create scene generator
	generator := scenegeneration.New(cfg.SceneGeneration)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Generate scenes
	scenes, err := generator.Generate(ctx, common.Content{})
	if err != nil {
		log.Fatalf("Failed to generate scenes: %v", err)
	}

	// Save scenes to file
	if err := generator.(interface {
		SaveScenesToFile([]common.Scene, string) error
	}).SaveScenesToFile(scenes, *outputPath); err != nil {
		log.Fatalf("Failed to save scenes to file: %v", err)
	}

	// Print success message
	log.Printf("Successfully generated %d scenes and saved to %s", len(scenes), *outputPath)
}
