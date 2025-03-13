# Stitch-Up Video Converters

This directory contains Node.js scripts for converting static images to videos using different AI models. These scripts are part of the Stitch-Up pipeline for generating news videos.

## Available Converters

### 1. RunwayML Video Converter (`video-converter.js`)

Uses RunwayML's Gen3 model to convert images to videos. This is the primary implementation that:
- Generates high-quality 10-second videos
- Uses RunwayML's Gen3 Turbo model
- Provides detailed motion and animation

### 2. Stable Video Diffusion via Replicate (`video-converter-replicate.js`)

An alternative implementation using Stability AI's Stable Video Diffusion model through Replicate:
- Generates 4-5 second videos
- Uses Stability AI's Stable Video Diffusion model
- Offers different motion characteristics and visual style
- Provides an alternative if RunwayML is unavailable

## Prerequisites

- Node.js 18+ (for native fetch API support)
- API keys for the services you want to use:
  - RunwayML API key for the primary converter
  - Replicate API key for the alternative converter

## Installation

1. Make sure you have Node.js installed
2. Install the required dependencies:

```bash
cd scripts
npm install
```

3. Create a `.env` file in the project root with your API keys:

```
RUNWAY_API_KEY=your_runway_api_key_here
REPLICATE_API_KEY=your_replicate_api_key_here
```

4. Make the scripts executable:

```bash
chmod +x scripts/video-converter.js
chmod +x scripts/video-converter-replicate.js
```

## Usage

### RunwayML Converter

```bash
# Convert all images in the default directory
node scripts/video-converter.js

# Convert a single image
node scripts/video-converter.js --input-dir path/to/image.png

# Specify custom directories
node scripts/video-converter.js --input-dir custom/images/dir --output-dir custom/videos/dir
```

### Replicate Converter

```bash
# Convert all images in the default directory
node scripts/video-converter-replicate.js

# Convert a single image
node scripts/video-converter-replicate.js --input-dir path/to/image.png

# Specify custom directories
node scripts/video-converter-replicate.js --input-dir custom/images/dir --output-dir custom/videos/dir

# Use a different Stable Video Diffusion model
node scripts/video-converter-replicate.js --model stability-ai/stable-video-diffusion-img2vid:9ca9f2b47f0e4f5c6b1e300aa0d5a961d22006f2a01058a2e93ddb1f9eb1d598
```

### Comparing Both Converters

A comparison script is provided to run both converters on the same image and compare the results:

```bash
# Run comparison with default image
node scripts/compare-video-converters.js

# Specify a custom image
node scripts/compare-video-converters.js --image path/to/image.png

# Specify a custom output directory
node scripts/compare-video-converters.js --output-dir custom/comparison/dir
```

## How It Works

Both converters follow a similar workflow:

1. Scan the input directory for image files
2. For each image:
   - Extract a scene ID from the filename
   - Look for matching scene information in `scenes.json`
   - Generate a rich description using scene details and mood
   - Send the image and description to the respective API
   - Poll for job completion
   - Download and save the generated video
3. Save metadata about all generated videos to `videos.json`

## Output

Both converters generate:

1. Video files in the output directory with names like `video_image_scene_name_hash_randomid.mp4`
2. A `videos.json` file containing metadata about each video:
   ```json
   [
     {
       "path": "output/videos/video_image_scene_name_hash_randomid.mp4",
       "imageID": "scene_name",
       "length": 10
     }
   ]
   ```

## Choosing Between Converters

- **RunwayML Converter**: Better for longer, more detailed videos with sophisticated motion
- **Replicate Converter**: Good alternative with different visual characteristics, potentially lower cost

## Troubleshooting

### RunwayML Converter Issues

- If you see 404 errors, check that your API key is valid and has access to Gen3 Turbo
- For timeout issues, the RunwayML API might be experiencing high load

### Replicate Converter Issues

- If you encounter API rate limits, try increasing the delay between requests
- For image format issues, ensure your images are in a supported format (PNG, JPG, etc.)

### General Issues

- Make sure your `.env` file contains the correct API keys
- Ensure you have Node.js 18+ installed
- Check that all dependencies are installed with `npm install` 