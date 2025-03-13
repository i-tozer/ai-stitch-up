# Video Converter using Stable Video Diffusion via Replicate

This script converts static images to videos using Stability AI's Stable Video Diffusion model through the Replicate API. It's designed as an alternative to the RunwayML-based video converter.

## Features

- Converts single images or directories of images to videos
- Extracts scene information from image filenames
- Looks up detailed scene descriptions from `scenes.json`
- Adds mood information to enhance video generation
- Handles errors gracefully with placeholder videos
- Saves metadata for downstream processing

## Prerequisites

- Node.js 18+ (for native fetch API support)
- A Replicate API key (sign up at [replicate.com](https://replicate.com))

## Installation

1. Make sure you have Node.js installed
2. Install the required dependencies:

```bash
npm install dotenv commander
```

3. Create a `.env` file in the project root with your Replicate API key:

```
REPLICATE_API_KEY=your_replicate_api_key_here
```

## Usage

### Basic Usage

Convert all images in the default directory:

```bash
node scripts/video-converter-replicate.js
```

### Options

- `-i, --input-dir <dir>`: Directory containing input images (default: `output/images`)
- `-o, --output-dir <dir>`: Directory for output videos (default: `output/videos`)
- `-l, --video-length <seconds>`: Length of generated videos in seconds (default: `10`)
- `-k, --api-key <key>`: Replicate API key (can also be set via REPLICATE_API_KEY env variable)
- `-m, --model <model>`: Replicate model to use (default: `stability-ai/stable-video-diffusion:3f0457e4619daac51203dedb472816fd4af51f3149fa7a9e0b5ffcf1b8172438`)

### Examples

Convert a single image:

```bash
node scripts/video-converter-replicate.js --input-dir path/to/image.png
```

Convert all images in a specific directory:

```bash
node scripts/video-converter-replicate.js --input-dir custom/images/dir --output-dir custom/videos/dir
```

Use a different Stable Video Diffusion model:

```bash
node scripts/video-converter-replicate.js --model stability-ai/stable-video-diffusion-img2vid:9ca9f2b47f0e4f5c6b1e300aa0d5a961d22006f2a01058a2e93ddb1f9eb1d598
```

## How It Works

1. The script scans the input directory for image files
2. For each image:
   - Extracts a scene ID from the filename
   - Looks for matching scene information in `scenes.json`
   - Generates a rich description using scene details and mood
   - Sends the image and description to Replicate's Stable Video Diffusion API
   - Polls for job completion
   - Downloads and saves the generated video
3. Saves metadata about all generated videos to `videos.json`

## Output

The script generates:

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

## Comparison with RunwayML Version

This implementation offers an alternative to the RunwayML-based video converter with some key differences:

- Uses Stability AI's Stable Video Diffusion model instead of RunwayML's Gen3 model
- Typically produces shorter videos (4-5 seconds vs. 10 seconds)
- May have different motion characteristics and visual style
- Uses a different pricing model (Replicate vs. RunwayML)

## Troubleshooting

- If you encounter API rate limits, try increasing the delay between requests
- For image format issues, ensure your images are in a supported format (PNG, JPG, etc.)
- Check the Replicate documentation for the latest model versions and parameters 