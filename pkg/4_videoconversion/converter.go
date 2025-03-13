package videoconversion

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

	"github.com/google/uuid"
	"github.com/iantozer/stitch-up/pkg/common"
	"github.com/iantozer/stitch-up/pkg/config"
)

// Converter implements the VideoConverter interface
type Converter struct {
	config config.VideoConversionConfig
	client *http.Client
}

// New creates a new video converter
func New(config config.VideoConversionConfig) common.VideoConverter {
	return NewConverter(config)
}

// Convert converts images to videos using Runway ML
func (c *Converter) Convert(ctx context.Context, images []common.Image) ([]common.Video, error) {
	log.Println("Converting images to videos using Runway ML")

	// Check if Runway API key is provided
	if c.config.RunwayAPIKey == "" {
		log.Println("No Runway API key provided, using placeholder videos")
		return c.createPlaceholderVideos(images)
	}

	var videos []common.Video

	for _, image := range images {
		log.Printf("Generating video for image: %s", image.Path)

		// Read the image file
		imageData, err := os.ReadFile(image.Path)
		if err != nil {
			log.Printf("Error reading image %s: %v", image.Path, err)
			continue
		}

		// Generate video using Runway ML
		videoData, err := c.generateVideoWithRunway(ctx, imageData, image.Description)
		if err != nil {
			log.Printf("Error generating video for image %s: %v", image.Path, err)
			continue
		}

		// Generate a unique filename
		baseFilename := filepath.Base(image.Path)
		baseFilename = strings.TrimSuffix(baseFilename, filepath.Ext(baseFilename))
		filename := fmt.Sprintf("video_%s_%s.mp4", baseFilename, uuid.New().String()[:8])

		videoPath := filepath.Join(c.config.OutputDir, filename)

		// Ensure the directory exists
		dir := filepath.Dir(videoPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Error creating directory for video %s: %v", videoPath, err)
			continue
		}

		// Save the video
		if err := os.WriteFile(videoPath, videoData, 0644); err != nil {
			log.Printf("Error saving video %s: %v", videoPath, err)
			continue
		}

		videos = append(videos, common.Video{
			Path:    videoPath,
			ImageID: image.SceneID,
			Length:  c.config.VideoLength,
		})

		log.Printf("Created video: %s", videoPath)

		// Add a small delay between API calls to avoid rate limiting
		time.Sleep(2 * time.Second)
	}

	if len(videos) == 0 {
		return videos, fmt.Errorf("no videos created")
	}

	log.Printf("Created %d videos", len(videos))
	return videos, nil
}

// generateVideoWithRunway generates a video from an image using Runway ML's API
func (c *Converter) generateVideoWithRunway(ctx context.Context, imageData []byte, description string) ([]byte, error) {
	// Runway ML API endpoint for image-to-video
	apiURL := "https://api.dev.runwayml.com/v1/image_to_video"

	// Encode the image as base64
	base64Image := c.encodeImageToBase64(imageData)

	// Prepare the request body with the new format
	requestBody := map[string]interface{}{
		"promptImage": base64Image,
		"promptText":  description,
		"model":       "gen3a_turbo",
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
	req.Header.Set("Authorization", "Bearer "+c.config.RunwayAPIKey)
	req.Header.Set("X-Runway-Version", "2024-11-06")

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
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Log the full response for debugging
	log.Printf("API Response: %s", string(body))

	// Parse response
	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors in the response
	if errMsg, ok := responseData["error"].(string); ok {
		return nil, fmt.Errorf("API error: %s", errMsg)
	}

	// Get the job ID from the response
	jobID, ok := responseData["id"].(string)
	if !ok {
		// Try alternative field names
		if jobID, ok = responseData["jobId"].(string); !ok {
			return nil, fmt.Errorf("no job ID in response: %v", responseData)
		}
	}

	log.Printf("Job ID: %s", jobID)

	// Poll for the result
	return c.pollForVideo(ctx, jobID)
}

// encodeImageToBase64 encodes an image as a base64 data URI
func (c *Converter) encodeImageToBase64(imageData []byte) string {
	// Determine the MIME type based on the image data
	mimeType := "image/jpeg" // Default to JPEG
	if len(imageData) > 2 {
		// Check for PNG signature
		if imageData[0] == 0x89 && imageData[1] == 0x50 && imageData[2] == 0x4E {
			mimeType = "image/png"
		}
	}

	// Encode the image data as base64
	base64Encoded := base64.StdEncoding.EncodeToString(imageData)

	// Return as a data URI
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Encoded)
}

// pollForVideo polls the Runway ML API for the generated video
func (c *Converter) pollForVideo(ctx context.Context, jobID string) ([]byte, error) {
	// Runway ML API endpoint for checking job status
	apiURL := fmt.Sprintf("https://api.dev.runwayml.com/v1/image_to_video/%s", jobID)

	log.Printf("Polling URL: %s", apiURL)

	// Maximum number of attempts
	maxAttempts := 60 // Videos can take longer to generate

	// Poll interval
	pollInterval := 5 * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Create HTTP request
		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("Authorization", "Bearer "+c.config.RunwayAPIKey)
		req.Header.Set("X-Runway-Version", "2024-11-06")

		// Log headers for debugging
		log.Printf("Request headers: %v", req.Header)

		// Send request
		resp, err := c.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to send request: %w", err)
		}

		// Read response
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		// Check response status
		if resp.StatusCode != http.StatusOK {
			log.Printf("Polling attempt %d: status code %d, body: %s", attempt, resp.StatusCode, string(body))

			// If we get a 404, the job ID might be in a different format or the endpoint is wrong
			if resp.StatusCode == http.StatusNotFound && attempt == 1 {
				// Try alternative polling URL format
				alternativeURL := fmt.Sprintf("https://api.dev.runwayml.com/v1/jobs/%s", jobID)
				log.Printf("Trying alternative polling URL: %s", alternativeURL)
				apiURL = alternativeURL
			}

			time.Sleep(pollInterval)
			continue
		}

		// Log the full response for debugging
		log.Printf("Polling response: %s", string(body))

		// Parse response
		var responseData map[string]interface{}
		if err := json.Unmarshal(body, &responseData); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		// Check the status of the generation
		status, ok := responseData["status"].(string)
		if !ok {
			log.Printf("Polling attempt %d: no status in response", attempt)
			time.Sleep(pollInterval)
			continue
		}

		if status == "completed" {
			// Get the video URL
			videoURL, ok := responseData["videoUrl"].(string)
			if !ok {
				// Try alternative field names
				if output, ok := responseData["output"].(map[string]interface{}); ok {
					if videoURL, ok = output["video"].(string); !ok {
						return nil, fmt.Errorf("no video URL in response: %v", responseData)
					}
				} else {
					return nil, fmt.Errorf("no video URL in response: %v", responseData)
				}
			}

			log.Printf("Video URL: %s", videoURL)

			// Download the video
			return c.downloadVideo(ctx, videoURL)
		}

		if status == "failed" {
			errorMessage := "unknown error"
			if errMsg, ok := responseData["error"].(string); ok {
				errorMessage = errMsg
			}
			return nil, fmt.Errorf("video generation failed: %s", errorMessage)
		}

		// Still processing, wait and try again
		log.Printf("Polling attempt %d: status %s", attempt, status)
		time.Sleep(pollInterval)
	}

	return nil, fmt.Errorf("timed out waiting for video generation")
}

// downloadVideo downloads a video from a URL
func (c *Converter) downloadVideo(ctx context.Context, url string) ([]byte, error) {
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response
	return io.ReadAll(resp.Body)
}

// createPlaceholderVideos creates placeholder videos for testing
func (c *Converter) createPlaceholderVideos(images []common.Image) ([]common.Video, error) {
	var videos []common.Video

	for _, image := range images {
		// Generate a unique filename
		baseFilename := filepath.Base(image.Path)
		baseFilename = strings.TrimSuffix(baseFilename, filepath.Ext(baseFilename))
		filename := fmt.Sprintf("placeholder_%s_%s.mp4", baseFilename, uuid.New().String()[:8])

		videoPath := filepath.Join(c.config.OutputDir, filename)

		// Create a placeholder video
		if err := createPlaceholderVideo(videoPath); err != nil {
			log.Printf("Error creating placeholder video %s: %v", videoPath, err)
			continue
		}

		videos = append(videos, common.Video{
			Path:    videoPath,
			ImageID: image.SceneID,
			Length:  c.config.VideoLength,
		})

		log.Printf("Created placeholder video: %s", videoPath)

		// Add a small delay to simulate API calls
		time.Sleep(100 * time.Millisecond)
	}

	return videos, nil
}

// createPlaceholderVideo creates an empty file as a placeholder
func createPlaceholderVideo(path string) error {
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
	_, err = file.WriteString("This is a placeholder for a video that would be generated by Runway ML")
	return err
}
