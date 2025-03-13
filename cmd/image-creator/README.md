# Image Creator

A command-line tool that generates images from scene descriptions using Hugging Face's API.

## Usage

1. Make sure you have generated scene descriptions using the scene-generator tool:
   ```
   ./bin/scene-generator
   ```

2. Set your Hugging Face API key in the `.env` file:
   ```
   HUGGINGFACE_API_KEY=your_huggingface_api_key_here
   ```

3. (Optional) Specify a different model to use:
   ```
   HUGGINGFACE_MODEL=runwayml/stable-diffusion-v1-5
   ```
   If not specified, it will default to `stabilityai/stable-diffusion-xl-base-1.0`.

4. Run the image creator:
   ```
   ./bin/image-creator
   ```

5. The generated images will be saved to `output/images` directory

## Options

- `--scenes`: Path to the scenes JSON file (default: `output/scenes.json`)
- `--output`: Directory to save the generated images (default: `output/images`)

## Example

```
./bin/image-creator --scenes output/bbc_scenes.json --output output/bbc_images
```

## How It Works

1. The tool loads scene descriptions from the JSON file
2. For each scene, it sends the description to Hugging Face's API
3. The API generates an image based on the scene description
4. The tool saves the generated image to the output directory
5. Image metadata is saved to `metadata.json` in the output directory

## Supported Models

You can use any image generation model available on Hugging Face. Here are some popular options:

- `stabilityai/stable-diffusion-xl-base-1.0` (default)
- `runwayml/stable-diffusion-v1-5`
- `CompVis/stable-diffusion-v1-4`
- `stabilityai/sdxl-turbo`
- `stabilityai/stable-diffusion-2-1`

For Stable Diffusion models, the tool automatically adds parameters like negative prompts and guidance scale to improve image quality 