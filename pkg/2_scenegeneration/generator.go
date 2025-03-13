/*
Package scenegeneration implements the second stage of the Stitch-Up pipeline.

This module is responsible for generating scene descriptions from BBC news content.
It takes a screenshot of the BBC website and uses Claude to analyze it and generate
visual scene descriptions that can be used for image creation.
*/
package scenegeneration

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/iantozer/stitch-up/pkg/common"
	"github.com/iantozer/stitch-up/pkg/config"
)

// Generator implements the SceneGenerator interface
type Generator struct {
	config config.SceneGenerationConfig
}

// New creates a new scene generator
func New(config config.SceneGenerationConfig) common.SceneGenerator {
	return &Generator{
		config: config,
	}
}

// Generate generates scene descriptions from BBC headline images
func (g *Generator) Generate(ctx context.Context, content common.Content) ([]common.Scene, error) {
	log.Println("Generating scene descriptions from BBC headline images")

	// Path to the BBC headline images directory
	imagesDir := "input/12_march_2025_bbc"

	// Check if the directory exists
	if _, err := os.Stat(imagesDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("images directory not found: %s", imagesDir)
	}

	// Read all files from the directory
	files, err := os.ReadDir(imagesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read images directory: %w", err)
	}

	// Filter image files
	var imageFiles []string
	for _, file := range files {
		if !file.IsDir() && isImageFile(file.Name()) {
			imageFiles = append(imageFiles, filepath.Join(imagesDir, file.Name()))
		}
	}

	if len(imageFiles) == 0 {
		return nil, fmt.Errorf("no image files found in directory: %s", imagesDir)
	}

	log.Printf("Found %d image files", len(imageFiles))

	// Process each image and generate a scene description
	var allScenes []common.Scene
	for _, imagePath := range imageFiles {
		log.Printf("Processing image: %s", imagePath)

		// Read the image
		imageData, err := os.ReadFile(imagePath)
		if err != nil {
			log.Printf("Warning: Failed to read image %s: %v", imagePath, err)
			continue
		}

		// Encode the image as base64
		base64Image := base64.StdEncoding.EncodeToString(imageData)

		// Generate scene description using Claude
		scenes, err := g.generateSceneForImage(ctx, base64Image, filepath.Base(imagePath))
		if err != nil {
			log.Printf("Warning: Failed to generate scene for image %s: %v", imagePath, err)
			continue
		}

		// Add the scene to the collection
		allScenes = append(allScenes, scenes...)
	}

	// If we couldn't generate any scenes, return mock scenes
	if len(allScenes) == 0 {
		log.Println("No scenes generated from images, using mock scenes")
		return g.getMockScenes(), nil
	}

	log.Printf("Generated %d scene descriptions", len(allScenes))
	return allScenes, nil
}

// isImageFile checks if a filename has an image extension
func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp"
}

// generateSceneForImage generates a scene description for a single image
func (g *Generator) generateSceneForImage(ctx context.Context, base64Image, imageName string) ([]common.Scene, error) {
	// Check if Claude API key is provided
	if g.config.ClaudeKey == "" {
		// Return a single mock scene for this image
		mockScenes := g.getMockScenes()
		if len(mockScenes) > 0 {
			return []common.Scene{mockScenes[0]}, nil
		}
		return nil, fmt.Errorf("no Claude API key provided and no mock scenes available")
	}

	// Prepare the prompt for Claude
	prompt := `You are an expert visual director. I'm showing you a screenshot of a BBC News headline.

Please analyze this news headline image and generate a single detailed scene description that visually represents this story.

Provide:
1. A title that captures the essence of the news story
2. A detailed visual description (150-200 words) that a text-to-image AI could use to generate a compelling image
3. The mood or atmosphere of the scene (e.g., tense, hopeful, somber)

Make the scene visually rich and emotionally impactful. Focus on creating imagery that tells the story without text.

Format your response as a JSON object with "title", "description", and "mood" fields.`

	// Call Claude API
	response, err := g.callClaudeAPI(ctx, prompt, base64Image)
	if err != nil {
		return nil, err
	}

	// Parse Claude's response
	scene, err := g.parseClaudeResponseForSingleScene(response, imageName)
	if err != nil {
		return nil, err
	}

	return []common.Scene{scene}, nil
}

// parseClaudeResponseForSingleScene parses Claude's response for a single scene
func (g *Generator) parseClaudeResponseForSingleScene(response, imageName string) (common.Scene, error) {
	// Extract JSON from response
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")

	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		// If we can't find valid JSON, try to extract scene description manually
		return g.extractSingleSceneManually(response, imageName)
	}

	jsonStr := response[jsonStart : jsonEnd+1]

	// Parse JSON
	var scene common.Scene
	if err := json.Unmarshal([]byte(jsonStr), &scene); err != nil {
		return g.extractSingleSceneManually(response, imageName)
	}

	// Add source information
	scene.SourceTitle = imageName
	scene.ID = generateSceneID(imageName)

	return scene, nil
}

// extractSingleSceneManually extracts a single scene description from Claude's response when JSON parsing fails
func (g *Generator) extractSingleSceneManually(response, imageName string) (common.Scene, error) {
	// Default values
	title := "News Scene: " + imageName
	description := "A visual representation of a news story."
	mood := "neutral"

	// Try to extract title
	titleStart := strings.Index(response, "Title:")
	if titleStart != -1 {
		titleEnd := strings.Index(response[titleStart:], "\n")
		if titleEnd != -1 {
			title = strings.TrimSpace(response[titleStart+6 : titleStart+titleEnd])
		}
	}

	// Try to extract description
	descStart := strings.Index(response, "Description:")
	if descStart != -1 {
		descEnd := strings.Index(response[descStart:], "Mood:")
		if descEnd != -1 {
			description = strings.TrimSpace(response[descStart+12 : descStart+descEnd])
		}
	}

	// Try to extract mood
	moodStart := strings.Index(response, "Mood:")
	if moodStart != -1 {
		moodEnd := strings.Index(response[moodStart:], "\n")
		if moodEnd != -1 {
			mood = strings.TrimSpace(response[moodStart+5 : moodStart+moodEnd])
		} else {
			mood = strings.TrimSpace(response[moodStart+5:])
		}
	}

	// If we couldn't extract anything meaningful, use the whole response as description
	if title == "News Scene: "+imageName && description == "A visual representation of a news story." {
		description = strings.TrimSpace(response)
	}

	return common.Scene{
		Title:       title,
		Description: description,
		Mood:        mood,
		SourceTitle: imageName,
		ID:          generateSceneID(imageName),
	}, nil
}

// generateSceneID generates a unique ID for a scene based on the image name
func generateSceneID(imageName string) string {
	// Remove extension
	name := strings.TrimSuffix(imageName, filepath.Ext(imageName))

	// Replace spaces and special characters
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")

	return "scene_" + name
}

// callClaudeAPI calls Claude's API with the prompt and image
func (g *Generator) callClaudeAPI(ctx context.Context, prompt, base64Image string) (string, error) {
	// Claude API endpoint
	apiURL := "https://api.anthropic.com/v1/messages"

	// Prepare the request body
	requestBody := map[string]interface{}{
		"model":      "claude-3-opus-20240229",
		"max_tokens": 4000,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": prompt,
					},
					{
						"type": "image",
						"source": map[string]string{
							"type":       "base64",
							"media_type": "image/png",
							"data":       base64Image,
						},
					},
				},
			},
		},
	}

	// Convert request body to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", g.config.ClaudeKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Send request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract content
	content, ok := responseData["content"].([]interface{})
	if !ok || len(content) == 0 {
		return "", fmt.Errorf("invalid response format")
	}

	// Get the text from the first content item
	contentItem, ok := content[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid content format")
	}

	text, ok := contentItem["text"].(string)
	if !ok {
		return "", fmt.Errorf("invalid text format")
	}

	return text, nil
}

// getMockScenes returns mock scene descriptions for testing
func (g *Generator) getMockScenes() []common.Scene {
	mockScenes := []common.Scene{
		{
			Title:       "Global Climate Summit",
			Description: "An aerial view of a massive climate conference center surrounded by protesters holding colorful signs. The building is modern with glass walls reflecting the crowd. In the foreground, world leaders in formal attire shake hands while environmental activists with painted faces watch from behind barriers. The sky shows dramatic clouds with rays of sunlight breaking through, symbolizing hope amid climate challenges.",
			Mood:        "Tense but hopeful",
		},
		{
			Title:       "Tech Regulation Debate",
			Description: "A split-screen visual showing two contrasting worlds. On the left, a bright, orderly digital landscape with transparent data flows and protected user information. On the right, a chaotic digital environment with shadowy figures extracting personal data. In the center, lawmakers stand at podiums debating while tech executives in sleek modern offices watch on screens. Digital privacy icons and regulation symbols float between the two worlds.",
			Mood:        "Confrontational",
		},
		{
			Title:       "Medical Breakthrough",
			Description: "A laboratory bathed in blue light where researchers in white coats examine 3D holographic brain scans showing Alzheimer's affected areas gradually healing. In the foreground, an elderly patient with a hopeful expression sits while a doctor explains results. The background shows microscopic views of the treatment targeting protein buildups. Family members watch through a glass window with expressions of cautious optimism.",
			Mood:        "Hopeful",
		},
		{
			Title:       "Global Economic Summit",
			Description: "A grand conference hall with representatives from major economies seated at a circular table. Digital displays show economic indicators and market trends. In the foreground, finance ministers exchange documents while advisors whisper in their ears. The background shows a world map with interconnected trade routes glowing in different intensities. Through large windows, protesters can be seen holding signs about economic inequality.",
			Mood:        "Tense",
		},
		{
			Title:       "Space Exploration Milestone",
			Description: "A control room erupting in celebration as screens show a spacecraft landing on a distant planet. Engineers and scientists embrace while others stare in awe at the main display showing the first images being transmitted back. The room is bathed in blue light from monitors, with national flags and mission patches adorning the walls. Through a large window, a night sky full of stars is visible, symbolizing humanity's ongoing journey into space.",
			Mood:        "Triumphant",
		},
		{
			Title:       "Refugee Crisis Response",
			Description: "A vast temporary settlement at sunset, with humanitarian workers distributing supplies. In the foreground, a doctor tends to a child while the family looks on with hope. Tents stretch to the horizon with people from diverse backgrounds helping each other. Aid helicopters hover overhead, delivering essential supplies. Despite the difficult circumstances, there are small moments of joy as children play and communities form.",
			Mood:        "Bittersweet",
		},
		{
			Title:       "Artificial Intelligence Breakthrough",
			Description: "A futuristic research lab where scientists interact with a holographic AI interface. The central AI visualization is represented as a complex, pulsating neural network in blue and purple hues. Researchers of diverse backgrounds collaborate around workstations while screens display complex equations and breakthrough results. Half the room shows practical applications: medical diagnostics, climate modeling, and language translation, while subtle lighting suggests both opportunity and caution.",
			Mood:        "Awe-inspiring",
		},
		{
			Title:       "Cultural Heritage Preservation",
			Description: "A team of conservationists and local community members carefully restoring an ancient temple damaged by climate events. Scaffolding surrounds parts of the structure while experts use both traditional methods and advanced technology to preserve intricate carvings. Elders share knowledge with younger generations, pointing to historical elements. The scene is bathed in warm golden light, with the surrounding landscape showing both environmental challenges and natural beauty.",
			Mood:        "Reverent",
		},
		{
			Title:       "Renewable Energy Revolution",
			Description: "A dramatic landscape transformation showing half the scene with traditional fossil fuel infrastructure and the other half with advanced renewable energy technology. Workers are seen transitioning from old to new industries. In the foreground, community members celebrate the opening of a solar farm while former coal workers receive training on wind turbine maintenance. The sky transitions from smoggy gray to clear blue across the image, with birds returning to the renewed environment.",
			Mood:        "Hopeful transition",
		},
		{
			Title:       "Global Health Initiative",
			Description: "A split scene showing healthcare workers delivering vaccines to remote villages alongside researchers in a state-of-the-art laboratory. The laboratory side shows scientists examining data while developing new treatments, while the field side shows the human impact of their work. A digital map in the center shows disease rates declining in real-time. People from diverse backgrounds and generations are shown both receiving care and contributing to the solution, emphasizing global cooperation.",
			Mood:        "Determined optimism",
		},
	}

	// Limit to max scenes
	if g.config.MaxScenes < len(mockScenes) {
		mockScenes = mockScenes[:g.config.MaxScenes]
	}

	return mockScenes
}

// SaveScenesToFile saves the generated scenes to a file
func (g *Generator) SaveScenesToFile(scenes []common.Scene, outputPath string) error {
	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Convert scenes to JSON
	jsonData, err := json.MarshalIndent(scenes, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scenes to JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write scenes to file: %w", err)
	}

	log.Printf("Saved %d scenes to %s", len(scenes), outputPath)
	return nil
}
