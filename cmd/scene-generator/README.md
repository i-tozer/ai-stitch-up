# Scene Generator

A command-line tool that processes individual BBC headline images and generates scene descriptions using Claude.

## Usage

1. Place your BBC headline images in the `input/12_march_2025_bbc` directory
   - Each image should represent a single news headline or story
   - Supported formats: jpg, jpeg, png, gif, webp

2. Set your Claude API key in the `.env` file:
   ```
   CLAUDE_API_KEY=your_claude_api_key_here
   ```

3. Run the scene generator:
   ```
   ./bin/scene-generator
   ```

4. The generated scenes will be saved to `output/scenes.json`

## Options

- `--output`: Path to save the generated scenes (default: `output/scenes.json`)
- `--max-scenes`: Maximum number of scenes to generate (default: 10)

## Example

```
./bin/scene-generator --max-scenes 15 --output output/bbc_scenes.json
```

## How It Works

1. The tool reads all image files from the `input/12_march_2025_bbc` directory
2. For each image, it sends the image to Claude with a prompt asking for a scene description
3. Claude analyzes each news headline image and generates a visual scene description
4. The tool parses Claude's responses and combines all scene descriptions
5. The combined scene descriptions are saved to a JSON file

Each scene includes:
- A title that captures the essence of the news story
- A detailed visual description that a text-to-image AI could use to generate an image
- The mood or atmosphere of the scene
- The source image filename
- A unique ID generated from the image filename 