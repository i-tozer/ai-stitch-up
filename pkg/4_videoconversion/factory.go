package videoconversion

import (
	"net/http"
	"time"

	"github.com/iantozer/stitch-up/pkg/common"
	"github.com/iantozer/stitch-up/pkg/config"
)

// NewConverter creates a new video converter based on the configuration
func NewConverter(config config.VideoConversionConfig) common.VideoConverter {
	// Check if we should use the Node.js implementation
	if config.UseNodeImplementation {
		return NewNodeWrapper(config)
	}

	// Fall back to the Go implementation
	return &Converter{
		config: config,
		client: &http.Client{
			Timeout: 120 * time.Second, // Longer timeout for video generation
		},
	}
}
