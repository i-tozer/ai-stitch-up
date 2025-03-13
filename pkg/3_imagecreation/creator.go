package imagecreation

// https://huggingface.co/docs/api-inference/en/getting-started

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/iantozer/stitch-up/pkg/common"
	"github.com/iantozer/stitch-up/pkg/config"
)

// Creator implements the ImageCreator interface
type Creator struct {
	config config.ImageCreationConfig
	client *http.Client
}

// New creates a new image creator
func New(config config.ImageCreationConfig) common.ImageCreator {
	return &Creator{
		config: config,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Create generates images from scene descriptions using Hugging Face's API
func (c *Creator) Create(ctx context.Context, scenes []common.Scene) ([]common.Image, error) {
	log.Println("Creating images from scene descriptions using Hugging Face's API")

	// Check if Hugging Face API key is provided
	if c.config.HuggingFaceAPIKey == "" {
		log.Println("No Hugging Face API key provided, using placeholder images")
		return c.createPlaceholderImages(scenes)
	}

	var images []common.Image

	for _, scene := range scenes {
		log.Printf("Generating image for scene: %s", scene.Title)

		// Generate image using Hugging Face's API
		imageData, err := c.generateImageWithHuggingFace(ctx, scene)
		if err != nil {
			log.Printf("Error generating image for scene %s: %v", scene.Title, err)
			continue
		}

		// Generate a unique filename
		filename := fmt.Sprintf("image_%s_%s.png",
			sanitizeFilename(scene.Title)[:20],
			uuid.New().String()[:8])

		imagePath := filepath.Join(c.config.OutputDir, filename)

		// Ensure the directory exists
		dir := filepath.Dir(imagePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Error creating directory for image %s: %v", imagePath, err)
			continue
		}

		// Save the image
		if err := os.WriteFile(imagePath, imageData, 0644); err != nil {
			log.Printf("Error saving image %s: %v", imagePath, err)
			continue
		}

		images = append(images, common.Image{
			Path:        imagePath,
			SceneID:     scene.ID,
			Description: scene.Description,
		})

		log.Printf("Created image: %s", imagePath)

		// Add a small delay between API calls to avoid rate limiting
		time.Sleep(2 * time.Second)
	}

	if len(images) == 0 {
		return images, fmt.Errorf("no images created")
	}

	log.Printf("Created %d images", len(images))
	return images, nil
}

// generateImageWithHuggingFace generates an image using Hugging Face's API
func (c *Creator) generateImageWithHuggingFace(ctx context.Context, scene common.Scene) ([]byte, error) {
	// Prepare the prompt
	prompt := c.preparePrompt(scene)

	// Hugging Face API endpoint for the specified model
	apiURL := fmt.Sprintf("https://api-inference.huggingface.co/models/%s", c.config.HuggingFaceModel)

	// Prepare the request body based on the model type
	var requestBody interface{}

	// Check if it's a Stable Diffusion model
	if strings.Contains(c.config.HuggingFaceModel, "stable-diffusion") {
		requestBody = map[string]interface{}{
			"inputs": prompt,
			"parameters": map[string]interface{}{
				"negative_prompt":     "blurry, low quality, distorted, deformed, disfigured",
				"num_inference_steps": 50,
				"guidance_scale":      7.5,
			},
		}
	} else {
		// Default request for other models
		requestBody = map[string]interface{}{
			"inputs": prompt,
		}
	}

	// Convert request body to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.HuggingFaceAPIKey)

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// For Hugging Face, the response is directly the image bytes for most image generation models
	// But some models might return JSON, so we need to check
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		// Try to parse as JSON
		var jsonResponse map[string]interface{}
		if err := json.Unmarshal(body, &jsonResponse); err == nil {
			// Check if there's an error message
			if errMsg, ok := jsonResponse["error"].(string); ok {
				return nil, fmt.Errorf("API error: %s", errMsg)
			}
		}
		return nil, fmt.Errorf("unexpected JSON response from image generation API")
	}

	// If we got here, the response should be the image bytes
	return body, nil
}

// preparePrompt prepares the prompt for Hugging Face's image generation API
func (c *Creator) preparePrompt(scene common.Scene) string {
	// Start with the scene description
	prompt := scene.Description

	// Add the mood if available
	if scene.Mood != "" {
		prompt += fmt.Sprintf(" The mood is %s.", scene.Mood)
	}

	// Add some style guidance based on the model
	if strings.Contains(c.config.HuggingFaceModel, "stable-diffusion") {
		prompt += " Photorealistic, high detail, dramatic lighting, 8k, cinematic, professional photography."
	}

	return prompt
}

// createPlaceholderImages creates placeholder images for testing
func (c *Creator) createPlaceholderImages(scenes []common.Scene) ([]common.Image, error) {
	var images []common.Image

	for _, scene := range scenes {
		// Generate a unique filename
		filename := fmt.Sprintf("placeholder_%s_%s.png",
			sanitizeFilename(scene.Title)[:20],
			uuid.New().String()[:8])

		imagePath := filepath.Join(c.config.OutputDir, filename)

		// Create a placeholder image
		if err := createPlaceholderImage(imagePath); err != nil {
			log.Printf("Error creating placeholder image %s: %v", imagePath, err)
			continue
		}

		images = append(images, common.Image{
			Path:        imagePath,
			SceneID:     scene.ID,
			Description: scene.Description,
		})

		log.Printf("Created placeholder image: %s", imagePath)

		// Add a small delay to simulate API calls
		time.Sleep(100 * time.Millisecond)
	}

	return images, nil
}

// createPlaceholderImage creates an empty file as a placeholder
func createPlaceholderImage(path string) error {
	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create an empty file
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write a placeholder message
	_, err = file.WriteString("This is a placeholder for an image that would be generated by Hugging Face's API")
	return err
}

// sanitizeFilename removes characters that are not allowed in filenames
func sanitizeFilename(filename string) string {
	// Replace spaces with underscores
	filename = strings.ReplaceAll(filename, " ", "_")

	// Remove special characters
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, ":", "_")
	filename = strings.ReplaceAll(filename, "*", "_")
	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "\"", "_")
	filename = strings.ReplaceAll(filename, "<", "_")
	filename = strings.ReplaceAll(filename, ">", "_")
	filename = strings.ReplaceAll(filename, "|", "_")

	// Convert to lowercase
	return strings.ToLower(filename)
}
