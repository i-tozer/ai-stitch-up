package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	videoconversion "github.com/iantozer/stitch-up/pkg/4_videoconversion"
	"github.com/iantozer/stitch-up/pkg/common"
	"github.com/iantozer/stitch-up/pkg/config"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	godotenv.Load()

	// Parse command line flags
	inputDir := flag.String("input-dir", "output/images", "Directory containing input images")
	outputDir := flag.String("output-dir", "output/videos", "Directory for output videos")
	videoLength := flag.Int("video-length", 10, "Length of generated videos in seconds")
	runwayAPIKey := flag.String("runway-api-key", os.Getenv("RUNWAY_API_KEY"), "Runway ML API key")
	useNode := flag.Bool("use-node", true, "Use Node.js implementation")
	flag.Parse()

	// Validate input
	if *runwayAPIKey == "" {
		log.Fatal("No Runway ML API key provided. Set RUNWAY_API_KEY in .env file or use --runway-api-key flag.")
	}

	// Create configuration
	cfg := config.VideoConversionConfig{
		RunwayAPIKey:          *runwayAPIKey,
		OutputDir:             *outputDir,
		VideoLength:           *videoLength,
		UseNodeImplementation: *useNode,
	}

	// Create video converter
	converter := videoconversion.New(cfg)

	// Get images from input directory
	images, err := getImagesFromDirectory(*inputDir)
	if err != nil {
		log.Fatalf("Failed to get images: %v", err)
	}

	if len(images) == 0 {
		log.Fatal("No images found in input directory")
	}

	log.Printf("Found %d images to convert", len(images))

	// Convert images to videos
	videos, err := converter.Convert(context.Background(), images)
	if err != nil {
		log.Fatalf("Failed to convert images to videos: %v", err)
	}

	log.Printf("Successfully converted %d images to videos", len(videos))
}

// getImagesFromDirectory gets all images from a directory
func getImagesFromDirectory(dir string) ([]common.Image, error) {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("input directory %s does not exist", dir)
	}

	// Read directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	// Filter image files
	var images []common.Image
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		ext := strings.ToLower(filepath.Ext(filename))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp" {
			// Extract scene ID from filename (assuming format: scene_<id>.<ext>)
			sceneID := strings.TrimSuffix(filename, ext)
			if strings.HasPrefix(sceneID, "scene_") {
				sceneID = strings.TrimPrefix(sceneID, "scene_")
			}

			images = append(images, common.Image{
				Path:    filepath.Join(dir, filename),
				SceneID: sceneID,
			})
		}
	}

	return images, nil
}
