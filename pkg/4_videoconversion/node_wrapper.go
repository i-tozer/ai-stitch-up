package videoconversion

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/iantozer/stitch-up/pkg/common"
	"github.com/iantozer/stitch-up/pkg/config"
)

// NodeWrapper implements the VideoConverter interface by wrapping the Node.js script
type NodeWrapper struct {
	config config.VideoConversionConfig
}

// NewNodeWrapper creates a new video converter that uses the Node.js script
func NewNodeWrapper(config config.VideoConversionConfig) common.VideoConverter {
	return &NodeWrapper{
		config: config,
	}
}

// Convert converts images to videos using the Node.js script
func (n *NodeWrapper) Convert(ctx context.Context, images []common.Image) ([]common.Video, error) {
	log.Println("Converting images to videos using Node.js script")

	// Check if Node.js is installed
	if err := checkNodeInstalled(); err != nil {
		return nil, fmt.Errorf("Node.js check failed: %w", err)
	}

	// Check if the script exists
	scriptPath, err := findScript()
	if err != nil {
		return nil, fmt.Errorf("script not found: %w", err)
	}

	// Check if npm dependencies are installed
	if err := checkDependencies(scriptPath); err != nil {
		return nil, fmt.Errorf("dependency check failed: %w", err)
	}

	// Prepare command arguments
	args := []string{
		scriptPath,
		"--input-dir", filepath.Dir(images[0].Path), // Assume all images are in the same directory
		"--output-dir", n.config.OutputDir,
		"--video-length", fmt.Sprintf("%d", n.config.VideoLength),
	}

	// Add API key if provided
	if n.config.RunwayAPIKey != "" {
		args = append(args, "--api-key", n.config.RunwayAPIKey)
	}

	// Create command
	cmd := exec.CommandContext(ctx, "node", args...)

	// Set environment variables
	cmd.Env = os.Environ()

	// Capture output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	log.Printf("Running command: node %s", strings.Join(args, " "))
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run Node.js script: %w", err)
	}

	// Read the output JSON file
	videos, err := readVideosJSON(filepath.Join(n.config.OutputDir, "videos.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to read videos JSON: %w", err)
	}

	log.Printf("Successfully converted %d images to videos", len(videos))
	return videos, nil
}

// checkNodeInstalled checks if Node.js is installed
func checkNodeInstalled() error {
	cmd := exec.Command("node", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Node.js is not installed: %w", err)
	}
	log.Printf("Node.js version: %s", strings.TrimSpace(string(output)))
	return nil
}

// findScript finds the path to the Node.js script
func findScript() (string, error) {
	// Check in the scripts directory
	scriptPath := filepath.Join("scripts", "video-converter.js")
	if _, err := os.Stat(scriptPath); err == nil {
		return scriptPath, nil
	}

	// Check in the current directory
	scriptPath = "video-converter.js"
	if _, err := os.Stat(scriptPath); err == nil {
		return scriptPath, nil
	}

	return "", fmt.Errorf("video-converter.js not found in scripts/ or current directory")
}

// checkDependencies checks if npm dependencies are installed
func checkDependencies(scriptPath string) error {
	scriptDir := filepath.Dir(scriptPath)

	// Check if node_modules exists
	nodeModulesPath := filepath.Join(scriptDir, "node_modules")
	if _, err := os.Stat(nodeModulesPath); os.IsNotExist(err) {
		log.Println("Installing dependencies...")

		// Run npm install
		cmd := exec.Command("npm", "install")
		cmd.Dir = scriptDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install dependencies: %w", err)
		}

		log.Println("Dependencies installed successfully")
	}

	return nil
}

// readVideosJSON reads the videos.json file and returns the videos
func readVideosJSON(path string) ([]common.Video, error) {
	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("videos.json file not found at %s", path)
	}

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read videos.json: %w", err)
	}

	// Define a struct to match the JSON format
	type videoJSON struct {
		Path    string `json:"path"`
		ImageID string `json:"imageID"`
		Length  int    `json:"length"`
	}

	// Parse the JSON
	var videosJSON []videoJSON
	if err := json.Unmarshal(data, &videosJSON); err != nil {
		return nil, fmt.Errorf("failed to parse videos.json: %w", err)
	}

	// Convert to common.Video
	videos := make([]common.Video, len(videosJSON))
	for i, v := range videosJSON {
		videos[i] = common.Video{
			Path:    v.Path,
			ImageID: v.ImageID,
			Length:  v.Length,
		}
	}

	return videos, nil
}
