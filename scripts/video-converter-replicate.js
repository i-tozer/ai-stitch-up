#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const dotenv = require('dotenv');
const { program } = require('commander');
const { promisify } = require('util');
const readdir = promisify(fs.readdir);
const readFile = promisify(fs.readFile);
const writeFile = promisify(fs.writeFile);
const mkdir = promisify(fs.mkdir);
const stat = promisify(fs.stat);

// Load environment variables from .env file
dotenv.config();

// Parse command line arguments
program
  .option('-i, --input-dir <dir>', 'Directory containing input images', 'output/images')
  .option('-o, --output-dir <dir>', 'Directory for output videos', 'output/videos')
  .option('-l, --video-length <seconds>', 'Length of generated videos in seconds', '10')
  .option('-k, --api-key <key>', 'Replicate API key', process.env.REPLICATE_API_KEY)
  .option('-m, --model <model>', 'Replicate model to use', 'stability-ai/stable-video-diffusion:3f0457e4619daac51203dedb472816fd4af51f3149fa7a9e0b5ffcf1b8172438')
  .parse(process.argv);

const options = program.opts();

// Validate options
if (!options.apiKey) {
  console.error('Error: No Replicate API key provided. Set REPLICATE_API_KEY in .env file or use --api-key flag.');
  process.exit(1);
}

// Check if a file is an image
function isImage(filename) {
  const ext = path.extname(filename).toLowerCase();
  return ['.jpg', '.jpeg', '.png', '.gif', '.webp'].includes(ext);
}

// Extract scene ID from filename
function extractSceneID(filename) {
  // The image filename format is typically "image_<scene_title>_<hash>.png"
  // We want to extract the scene title part
  const basename = path.basename(filename, path.extname(filename));
  
  // Try to match the pattern "image_<scene_title>_<hash>"
  const match = basename.match(/^image_(.+)_[a-z0-9]+$/);
  if (match && match[1]) {
    return match[1]; // Return the scene title part
  }
  
  // If no match, return the basename as fallback
  return basename;
}

// Find matching scene in scenes.json
async function findMatchingScene(sceneID) {
  try {
    const scenesPath = path.join('output', 'scenes.json');
    if (!fs.existsSync(scenesPath)) {
      console.log('scenes.json not found');
      return null;
    }
    
    const scenes = JSON.parse(await readFile(scenesPath, 'utf8'));
    
    // First try to find an exact match by ID
    let scene = scenes.find(s => s.id === sceneID);
    
    // If no exact match, try to find a scene with a title that contains the sceneID
    if (!scene) {
      scene = scenes.find(s => {
        const title = (s.title || '').toLowerCase();
        const sourceTitle = (s.source_title || '').toLowerCase();
        const searchID = sceneID.toLowerCase().replace(/_/g, ' ');
        
        return title.includes(searchID) || sourceTitle.includes(searchID);
      });
    }
    
    return scene;
  } catch (err) {
    console.log(`Error finding matching scene: ${err.message}`);
    return null;
  }
}

// Download a file from a URL
async function downloadFile(url) {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Failed to download file: ${response.statusText}`);
  }
  return Buffer.from(await response.arrayBuffer());
}

// Create a placeholder video file
async function createPlaceholderVideo(filePath) {
  const dir = path.dirname(filePath);
  await mkdir(dir, { recursive: true });
  await writeFile(filePath, 'This is a placeholder for a video that would be generated by Stable Video Diffusion');
  return filePath;
}

// Get description for an image
async function getDescriptionForImage(imagePath, sceneID) {
  let description = `Generated from scene ${sceneID}`;
  
  try {
    // Look for a metadata file with the same name but .json extension
    const metadataPath = path.join(path.dirname(imagePath), `${path.basename(imagePath, path.extname(imagePath))}.json`);
    if (fs.existsSync(metadataPath)) {
      const metadata = JSON.parse(await readFile(metadataPath, 'utf8'));
      if (metadata.description) {
        description = metadata.description;
        console.log(`Using description from metadata file: "${description}"`);
        return description;
      }
    }
    
    // Try to find the scene in scenes.json
    const scene = await findMatchingScene(sceneID);
    if (scene) {
      // Use the scene description if available, otherwise use the title
      if (scene.description) {
        description = scene.description;
        console.log(`Using description from scenes.json: "${description}"`);
      } else if (scene.title) {
        description = `A scene depicting ${scene.title}`;
        console.log(`Using title from scenes.json: "${description}"`);
      }
      
      // Add mood if available
      if (scene.mood) {
        description += `. The mood is ${scene.mood}.`;
        console.log(`Added mood to description: "${description}"`);
      }
      
      return description;
    }
  } catch (err) {
    console.log(`Could not load metadata, using default description: ${err.message}`);
  }
  
  return description;
}

// Generate video using Replicate's Stable Video Diffusion
async function generateVideoWithReplicate(imageData, description, model) {
  console.log(`Generating video using Replicate model: ${model}`);
  
  // Convert image to base64
  const base64Image = `data:image/png;base64,${imageData.toString('base64')}`;
  
  // Create a prediction with Replicate API
  const createResponse = await fetch("https://api.replicate.com/v1/predictions", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Authorization": `Token ${options.apiKey}`
    },
    body: JSON.stringify({
      version: model.split(':')[1] || model,
      input: {
        image: base64Image,
        prompt: description,
        motion_bucket_id: 127,
        fps: 6,
        num_frames: 25,
        noise_aug_strength: 0.1
      }
    })
  });
  
  if (!createResponse.ok) {
    const errorText = await createResponse.text();
    throw new Error(`Replicate API error (${createResponse.status}): ${errorText}`);
  }
  
  const prediction = await createResponse.json();
  console.log(`Prediction created with ID: ${prediction.id}`);
  
  // Poll for completion
  let result;
  while (true) {
    console.log('Waiting 5 seconds before checking status...');
    await new Promise(resolve => setTimeout(resolve, 5000));
    
    const statusResponse = await fetch(`https://api.replicate.com/v1/predictions/${prediction.id}`, {
      headers: {
        "Authorization": `Token ${options.apiKey}`
      }
    });
    
    if (!statusResponse.ok) {
      const errorText = await statusResponse.text();
      throw new Error(`Replicate API error (${statusResponse.status}): ${errorText}`);
    }
    
    result = await statusResponse.json();
    console.log(`Current status: ${result.status}`);
    
    if (result.status === "succeeded") {
      break;
    } else if (result.status === "failed") {
      throw new Error(`Prediction failed: ${result.error || "Unknown error"}`);
    }
  }
  
  // Get the video URL from the output
  const videoUrl = result.output;
  if (!videoUrl) {
    throw new Error("No video URL in the prediction output");
  }
  
  console.log(`Downloading video from: ${videoUrl}`);
  return await downloadFile(videoUrl);
}

// Process a single image
async function processImage(imageFile, inputDir, outputDir, videoLength) {
  const imagePath = path.join(inputDir, imageFile);
  const sceneID = extractSceneID(imageFile);
  
  console.log(`Processing image: ${imagePath}`);
  
  try {
    // Read the image file
    const imageData = await readFile(imagePath);
    
    // Get description for the image
    const description = await getDescriptionForImage(imagePath, sceneID);
    
    // Generate video using Replicate
    console.log(`Creating video for image: ${imageFile}`);
    console.log(`Using description: "${description}"`);
    
    // Generate the video
    console.log('Sending request to Replicate...');
    const videoData = await generateVideoWithReplicate(imageData, description, options.model);
    console.log('Video generation completed');
    
    // Generate output filename
    const baseFilename = path.basename(imageFile, path.extname(imageFile));
    const randomId = Math.random().toString(36).substring(2, 10);
    const outputFilename = `video_${baseFilename}_${randomId}.mp4`;
    const outputPath = path.join(outputDir, outputFilename);
    
    // Save the video
    console.log(`Saving video to: ${outputPath}`);
    await writeFile(outputPath, videoData);
    
    return {
      path: outputPath,
      imageID: sceneID,
      length: parseInt(videoLength, 10)
    };
  } catch (error) {
    console.error(`Error processing image ${imageFile}:`, error);
    
    // Create a placeholder video if real generation fails
    const baseFilename = path.basename(imageFile, path.extname(imageFile));
    const randomId = Math.random().toString(36).substring(2, 10);
    const outputFilename = `placeholder_${baseFilename}_${randomId}.mp4`;
    const outputPath = path.join(outputDir, outputFilename);
    
    await createPlaceholderVideo(outputPath);
    
    console.log(`Created placeholder video: ${outputPath}`);
    
    return {
      path: outputPath,
      imageID: sceneID,
      length: parseInt(videoLength, 10)
    };
  }
}

// Main function to convert images to videos
async function convertImagesToVideos() {
  console.log(`Starting video conversion at ${new Date().toLocaleTimeString()}`);
  console.log(`Input directory: ${options.inputDir}`);
  console.log(`Output directory: ${options.outputDir}`);
  console.log(`Using Replicate model: ${options.model}`);
  
  try {
    // Check if input is a directory or a single file
    const inputStats = await stat(options.inputDir);
    let imageFiles = [];
    
    if (inputStats.isDirectory()) {
      // It's a directory, get all image files
      const files = await readdir(options.inputDir);
      imageFiles = files.filter(file => isImage(file));
      
      if (imageFiles.length === 0) {
        console.error(`Error: No image files found in ${options.inputDir}`);
        process.exit(1);
      }
    } else if (inputStats.isFile() && isImage(options.inputDir)) {
      // It's a single image file
      imageFiles = [path.basename(options.inputDir)];
      options.inputDir = path.dirname(options.inputDir);
    } else {
      console.error(`Error: ${options.inputDir} is not a directory or an image file`);
      process.exit(1);
    }
    
    // Create output directory if it doesn't exist
    await mkdir(options.outputDir, { recursive: true });
    
    console.log(`Found ${imageFiles.length} images to convert`);
    
    // Process each image
    const results = [];
    for (const imageFile of imageFiles) {
      const result = await processImage(imageFile, options.inputDir, options.outputDir, options.videoLength);
      results.push(result);
      
      // Add a small delay between API calls to avoid rate limiting
      await new Promise(resolve => setTimeout(resolve, 2000));
    }
    
    console.log(`Video conversion completed at ${new Date().toLocaleTimeString()}`);
    console.log(`Generated ${results.length} videos in ${options.outputDir}`);
    
    // Save results to a JSON file for further processing
    const resultsPath = path.join(options.outputDir, 'videos.json');
    await writeFile(resultsPath, JSON.stringify(results, null, 2));
    console.log(`Results saved to: ${resultsPath}`);
    
  } catch (error) {
    console.error('Error during video conversion:', error);
    process.exit(1);
  }
}

// Run the main function
convertImagesToVideos().catch(error => {
  console.error('Unhandled error:', error);
  process.exit(1);
}); 