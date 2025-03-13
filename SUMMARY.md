# Stitch-Up Project Summary

## What We've Built

We've created a simple scene generator that takes a screenshot of the BBC website and generates visual scene descriptions using Claude. The key components are:

1. **Scene Generator Module** (`pkg/2_scenegeneration/generator.go`):
   - Takes a BBC screenshot as input
   - Sends it to Claude with a prompt asking for scene descriptions
   - Parses Claude's response into structured scene descriptions
   - Includes fallback mechanisms for when Claude's API is unavailable

2. **Command-Line Tool** (`cmd/scene-generator/main.go`):
   - Simple CLI tool to run the scene generator
   - Configurable output path and maximum number of scenes
   - Saves the generated scenes to a JSON file

3. **Helper Script** (`scripts/generate_scenes.sh`):
   - Loads environment variables from `.env`
   - Checks for the BBC screenshot
   - Builds and runs the scene generator
   - Provides helpful error messages

## How to Use

1. Place a screenshot of the BBC website in `input/12_march_2025_bbc.png`
2. Set your Claude API key in the `.env` file
3. Run the script: `./scripts/generate_scenes.sh`
4. The generated scenes will be saved to `output/scenes.json`

## Next Steps

The scene generator is just one part of the Stitch-Up pipeline. The next steps would be to:

1. Use the generated scene descriptions to create images with Ideogram AI
2. Convert those images to videos with Runway ML
3. Generate lyrics based on the news content with Claude
4. Create music from the lyrics with Suno AI
5. Combine everything into a final video with ffmpeg

## Design Decisions

1. **Simplicity**: We kept the design very simple - one input, one prompt, one output.
2. **Resilience**: We added fallback mechanisms for when Claude's API is unavailable.
3. **Configurability**: The scene generator can be configured via command-line flags and environment variables.
4. **Modularity**: The scene generator is just one module in the larger Stitch-Up pipeline. 