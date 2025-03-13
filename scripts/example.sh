#!/bin/bash

# Example script to demonstrate the simple video generator

# Set the path to an example image
IMAGE_PATH="output/images/image_bridgerton_vs._stric_09fba2da.png"

# Create output directory if it doesn't exist
mkdir -p output/videos/examples

# Example 1: Basic usage with a simple description
echo "Example 1: Basic usage with a simple description"
node scripts/simple-video-generator.js \
  --input "$IMAGE_PATH" \
  --description "A grand ballroom with contrasting dance styles" \
  --output "output/videos/examples/example1.mp4"

echo ""

# Example 2: Detailed description with motion cues
echo "Example 2: Detailed description with motion cues"
node scripts/simple-video-generator.js \
  --input "$IMAGE_PATH" \
  --description "A grand ballroom scene with two contrasting dance styles: on one side, elegant Regency-era dancers in formal wear performing a classical waltz with graceful spinning movements; on the other side, modern dancers in contemporary clothing performing energetic street dance moves with dynamic jumps and spins. The camera slowly pans across the scene, highlighting the dramatic contrast." \
  --output "output/videos/examples/example2.mp4"

echo ""

# Example 3: Description with mood and atmosphere
echo "Example 3: Description with mood and atmosphere"
node scripts/simple-video-generator.js \
  --input "$IMAGE_PATH" \
  --description "A grand ballroom with contrasting dance styles. The mood is dramatic, enchanting, and full of energy. The lighting shifts between warm golden tones on the classical dancers and cool blue tones on the modern dancers, creating a visual tension between tradition and innovation." \
  --output "output/videos/examples/example3.mp4"

echo ""

echo "All examples completed. Videos saved to output/videos/examples/"
echo "Compare the different videos to see how the description affects the generated video." 