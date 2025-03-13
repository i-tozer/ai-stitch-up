package lyriccreation

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/iantozer/stitch-up/pkg/common"
	"github.com/iantozer/stitch-up/pkg/config"
)

func TestCreator_Create(t *testing.T) {
	// Create test configuration
	cfg := config.LyricCreationConfig{}

	// Create creator instance
	creator := New(cfg)

	// Create test content
	content := common.Content{
		Title:       "Test News",
		Description: "Test Description",
		Date:        time.Now().Format("January 2, 2006"),
		Articles: []common.Article{
			{
				Title:       "Market Update",
				Summary:     "Market summary",
				Content:     "Market content",
				URL:         "https://example.com/market",
				PublishedAt: time.Now().Format(time.RFC3339),
			},
			{
				Title:       "Technology News",
				Summary:     "Tech summary",
				Content:     "Tech content",
				URL:         "https://example.com/tech",
				PublishedAt: time.Now().Format(time.RFC3339),
			},
		},
	}

	// Test lyric creation
	ctx := context.Background()
	lyrics, err := creator.Create(ctx, content)
	if err != nil {
		t.Errorf("Create() error = %v", err)
		return
	}

	// Validate lyrics structure
	if lyrics.Title == "" {
		t.Error("Create() returned empty title")
	}
	if !strings.Contains(lyrics.Title, content.Date) {
		t.Errorf("Lyrics title does not contain date: %s", lyrics.Title)
	}

	// Validate lyrics content
	if lyrics.Content == "" {
		t.Error("Create() returned empty content")
	}

	// Check for expected song structure
	expectedSections := []string{"VERSE", "CHORUS", "BRIDGE"}
	for _, section := range expectedSections {
		if !strings.Contains(lyrics.Content, section) {
			t.Errorf("Lyrics missing %s section", section)
		}
	}

	// Check for reasonable line count
	lines := strings.Split(lyrics.Content, "\n")
	if len(lines) < 10 {
		t.Error("Lyrics seem too short")
	}
}

func TestCreator_Create_EmptyContent(t *testing.T) {
	// Test with empty content
	cfg := config.LyricCreationConfig{}
	creator := New(cfg)
	ctx := context.Background()
	content := common.Content{
		Title:       "Empty Test",
		Description: "Empty Description",
		Date:        time.Now().Format("January 2, 2006"),
		Articles:    []common.Article{}, // Empty articles
	}

	lyrics, err := creator.Create(ctx, content)
	// The current implementation still generates placeholder lyrics
	// In a real implementation, this might return an error
	if err != nil {
		t.Errorf("Create() with empty content error = %v", err)
	}
	if lyrics.Title == "" || lyrics.Content == "" {
		t.Error("Create() with empty content returned empty lyrics")
	}
}

func TestCreator_Create_ContentValidation(t *testing.T) {
	// Test with invalid content
	cfg := config.LyricCreationConfig{}
	creator := New(cfg)
	ctx := context.Background()
	content := common.Content{
		Title:       "",             // Empty title
		Description: "",             // Empty description
		Date:        "invalid date", // Invalid date
		Articles:    nil,            // Nil articles
	}

	lyrics, err := creator.Create(ctx, content)
	// The current implementation still generates placeholder lyrics
	// In a real implementation, this might validate the content more strictly
	if err != nil {
		t.Errorf("Create() with invalid content error = %v", err)
	}
	if lyrics.Title == "" || lyrics.Content == "" {
		t.Error("Create() with invalid content returned empty lyrics")
	}
}
