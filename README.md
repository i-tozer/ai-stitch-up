# Stitch-Up

A pipeline for generating news videos with AI.

https://dev.runwayml.com/organization/fedf391b-43ca-411c-bbd5-12c05defd6a1/api-keys

## Overview
1. **Content Extraction**: Extract news content from the BBC website
2. **Scene Generation**: Generate visual scene descriptions from BBC screenshots
3. **Image Creation**: Create images from scene descriptions using HuggingFace
4. **Video Conversion**: Convert images to videos using Runway ML
5. **Lyric Creation**: Generate lyrics based on the news content using Claude
6. **Music Generation**: Create music from the lyrics using Suno AI
7. **Final Assembly**: Combine videos and music into a final presentation using ffmpeg

## Scene Generator

The Scene Generator is a simple tool that takes a screenshot of the BBC website and generates visual scene descriptions using Claude.

### Usage

1. Place a screenshot of the BBC website in `input/12_march_2025_bbc.png`

2. Set your Claude API key in the `.env` file:
   ```
   CLAUDE_API_KEY=your_claude_api_key_here
   ```

3. Run the scene generator:
   ```
   ./bin/scene-generator
   ```

4. The generated scenes will be saved to `output/scenes.json`

For more details, see the [Scene Generator README](cmd/scene-generator/README.md).

## Environment Setup

Stitch-Up uses environment variables for configuration. You can set these in a `.env` file in the project root.

1. Copy the example environment file:
   ```
   cp .env.example .env
   ```

2. Edit the `.env` file and add your API keys:
   ```
   CLAUDE_API_KEY=your_claude_api_key_here
   ```

## Configuration

The following environment variables can be set in your `.env` file:

| Variable | Description |
|----------|-------------|
| `CLAUDE_API_KEY` | API key for Claude AI |
| `BBC_URL` | URL for BBC news (default: https://www.bbc.com/news) |
| `OUTPUT_DIR` | Directory for output files (default: ./output) |
| `IDEOGRAM_API_KEY` | API key for Ideogram |
| `RUNWAY_API_KEY` | API key for Runway |
| `SUNO_API_KEY` | API key for Suno |
| `REAL_TEST` | Set to "true" to run tests against the real BBC website |

## Project Structure

The project is organized into modules that represent each stage of the pipeline:

1. Content Extraction (`pkg/1_contentextraction`)
2. Scene Generation (`pkg/2_scenegeneration`)
3. Image Creation (`pkg/3_imagecreation`)
4. Video Conversion (`pkg/4_videoconversion`)
5. Lyric Creation (`pkg/5_lyriccreation`)
6. Music Generation (`pkg/6_musicgeneration`)
7. Assembly (`pkg/7_assembly`)