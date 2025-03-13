package main

import (
	"context"
	"fmt"
	"log"
	"os"

	// contentextraction "github.com/iantozer/stitch-up/pkg/1_contentextraction"
	scenegeneration "github.com/iantozer/stitch-up/pkg/2_scenegeneration"
	imagecreation "github.com/iantozer/stitch-up/pkg/3_imagecreation"
	videoconversion "github.com/iantozer/stitch-up/pkg/4_videoconversion"
	lyriccreation "github.com/iantozer/stitch-up/pkg/5_lyriccreation"
	musicgeneration "github.com/iantozer/stitch-up/pkg/6_musicgeneration"
	assembly "github.com/iantozer/stitch-up/pkg/7_assembly"
	"github.com/iantozer/stitch-up/pkg/common"
	"github.com/iantozer/stitch-up/pkg/config"
)

func main() {
	// Initialize context
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize modules
	// contentExtractor := contentextraction.New(cfg.ContentExtraction)
	sceneGenerator := scenegeneration.New(cfg.SceneGeneration)
	imageCreator := imagecreation.New(cfg.ImageCreation)
	videoConverter := videoconversion.New(cfg.VideoConversion)
	lyricCreator := lyriccreation.New(cfg.LyricCreation)
	musicGenerator := musicgeneration.New(cfg.MusicGeneration)
	assembler := assembly.New(cfg.Assembly)

	// Extract content
	fmt.Println("Step 1: Extracting content...")
	// content, err := contentExtractor.Extract(ctx)
	// if err != nil {
	// 	log.Fatalf("Content extraction failed: %v", err)
	// }

	content := common.Content{
		Title: "Test Content",
		// Body:  "This is a test content body",
	}

	// Generate scene descriptions
	fmt.Println("Step 2: Generating scene descriptions...")
	scenes, err := sceneGenerator.Generate(ctx, content)
	if err != nil {
		log.Fatalf("Scene generation failed: %v", err)
	}

	// Create images from scene descriptions
	fmt.Println("Step 3: Creating images...")
	images, err := imageCreator.Create(ctx, scenes)
	if err != nil {
		log.Fatalf("Image creation failed: %v", err)
	}

	// Convert images to videos
	fmt.Println("Step 4: Converting images to videos...")
	videos, err := videoConverter.Convert(ctx, images)
	if err != nil {
		log.Fatalf("Video conversion failed: %v", err)
	}

	// Create lyrics based on content
	fmt.Println("Step 5: Creating lyrics...")
	lyrics, err := lyricCreator.Create(ctx, content)
	if err != nil {
		log.Fatalf("Lyric creation failed: %v", err)
	}

	// Generate music from lyrics
	fmt.Println("Step 6: Generating music...")
	music, err := musicGenerator.Generate(ctx, lyrics)
	if err != nil {
		log.Fatalf("Music generation failed: %v", err)
	}

	// Assemble final output
	fmt.Println("Step 7: Assembling final output...")
	outputPath, err := assembler.Assemble(ctx, videos, music)
	if err != nil {
		log.Fatalf("Assembly failed: %v", err)
	}

	fmt.Printf("Process completed successfully! Output saved to: %s\n", outputPath)
	os.Exit(0)
}
