package common

import (
	"context"
)

// Content represents extracted news content
type Content struct {
	Title       string
	Description string
	Articles    []Article
	Date        string
}

// Article represents a single news article
type Article struct {
	Title       string
	Summary     string
	Content     string
	URL         string
	PublishedAt string
}

// Scene represents a visual scene description
type Scene struct {
	Description string `json:"description"`
	SourceTitle string `json:"source_title"`
	ID          string `json:"id"`
	Title       string `json:"title"`
	Mood        string `json:"mood"`
}

// Image represents a generated image
type Image struct {
	Path        string
	SceneID     string
	Description string
}

// Video represents a generated video clip
type Video struct {
	Path    string
	ImageID string
	Length  int // in seconds
}

// Lyrics represents generated song lyrics
type Lyrics struct {
	Title   string
	Content string
}

// Music represents a generated music track
type Music struct {
	Path     string
	LyricsID string
	Length   int // in seconds
}

// ContentExtractor extracts news content
type ContentExtractor interface {
	Extract(ctx context.Context) (Content, error)
}

// SceneGenerator generates scene descriptions
type SceneGenerator interface {
	Generate(ctx context.Context, content Content) ([]Scene, error)
}

// ImageCreator creates images from scene descriptions
type ImageCreator interface {
	Create(ctx context.Context, scenes []Scene) ([]Image, error)
}

// VideoConverter converts images to videos
type VideoConverter interface {
	Convert(ctx context.Context, images []Image) ([]Video, error)
}

// LyricCreator creates lyrics from content
type LyricCreator interface {
	Create(ctx context.Context, content Content) (Lyrics, error)
}

// MusicGenerator generates music from lyrics
type MusicGenerator interface {
	Generate(ctx context.Context, lyrics Lyrics) (Music, error)
}

// Assembler assembles videos and music into final output
type Assembler interface {
	Assemble(ctx context.Context, videos []Video, music Music) (string, error)
}
