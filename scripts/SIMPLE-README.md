# Simple Video Generator

A streamlined Node.js script that generates a video from a single image and description using RunwayML's Gen3 Turbo model.

## Features

- Simple command-line interface with required parameters
- Generates a 5-second video (customizable)
- Uses RunwayML's Gen3 Turbo model for high-quality results
- Detailed logging of the generation process

## Prerequisites

- Node.js 18+ (for native fetch API support)
- RunwayML API key

## Installation

1. Make sure you have Node.js installed
2. Install the required dependencies:

```bash
npm install dotenv commander @runwayml/sdk
```

3. Create a `.env` file in the project root with your RunwayML API key:

```
RUNWAY_API_KEY=your_runway_api_key_here
```

4. Make the script executable:

```bash
chmod +x scripts/simple-video-generator.js
```

## Usage

```bash
node scripts/simple-video-generator.js --input <image-path> --description "<description-text>" --output <output-path>
```

### Required Parameters

- `-i, --input <path>`: Path to the input image
- `-d, --description <text>`: Description of the image for video generation
- `-o, --output <path>`: Path where the output video will be saved

### Optional Parameters

- `-k, --api-key <key>`: RunwayML API key (can also be set via RUNWAY_API_KEY env variable)
- `-l, --length <seconds>`: Length of the video in seconds (default: 5)

### Examples

Generate a video with a specific description:

```bash
node scripts/simple-video-generator.js \
  --input path/to/image.png \
  --description "A serene lake with mountains in the background, gentle ripples on the water surface, birds flying overhead" \
  --output path/to/output.mp4
```

Specify a custom API key:

```bash
node scripts/simple-video-generator.js \
  --input path/to/image.png \
  --description "Description text" \
  --output path/to/output.mp4 \
  --api-key your-api-key-here
```

## How It Works

1. The script reads the input image file
2. Converts the image to a base64-encoded data URI
3. Sends the image and description to RunwayML's image-to-video API
4. Polls for job completion
5. Downloads and saves the generated video

## Tips for Good Results

1. **Detailed Descriptions**: The more detailed your description, the better the video will be. Include:
   - What elements should move (e.g., "waves crashing", "leaves rustling")
   - The mood or atmosphere (e.g., "peaceful", "tense", "joyful")
   - Any specific motion directions (e.g., "camera slowly panning right")

2. **Image Quality**: Use high-quality images with clear subjects and good lighting

3. **Experiment**: Try different descriptions to see how they affect the generated video 