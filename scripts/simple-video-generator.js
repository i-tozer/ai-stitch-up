#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const RunwayML = require('@runwayml/sdk');
const dotenv = require('dotenv');
const { program } = require('commander');
const { promisify } = require('util');
const readFile = promisify(fs.readFile);
const writeFile = promisify(fs.writeFile);
const mkdir = promisify(fs.mkdir);

// Load environment variables from .env file
dotenv.config();

// Parse command line arguments
program
  .requiredOption('-i, --input <path>', 'Path to the input image')
  .requiredOption('-d, --description <text>', 'Description of the image for video generation')
  .requiredOption('-o, --output <path>', 'Path where the output video will be saved')
  .option('-k, --api-key <key>', 'RunwayML API key', process.env.RUNWAY_API_KEY)
  .option('-l, --length <seconds>', 'Length of the video in seconds', '5')
  .parse(process.argv);

const options = program.opts();

// Validate options
if (!options.apiKey) {
  console.error('Error: No RunwayML API key provided. Set RUNWAY_API_KEY in .env file or use --api-key flag.');
  process.exit(1);
}

// Initialize the RunwayML client
const client = new RunwayML({
  apiKey: options.apiKey,
});

// Download a file from a URL
async function downloadFile(url) {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Failed to download file: ${response.statusText}`);
  }
  return Buffer.from(await response.arrayBuffer());
}

// Main function to generate video
async function generateVideo() {
  console.log(`Starting video generation at ${new Date().toLocaleTimeString()}`);
  console.log(`Input image: ${options.input}`);
  console.log(`Description: ${options.description}`);
  console.log(`Output path: ${options.output}`);
  console.log(`Video length: ${options.length} seconds`);
  
  try {
    // Read the image file
    console.log('Reading input image...');
    const imageData = await readFile(options.input);
    
    // Convert image to base64
    const imageExt = path.extname(options.input).substring(1).toLowerCase();
    const mimeType = imageExt === 'jpg' ? 'jpeg' : imageExt; // Handle jpg extension
    const base64Image = `data:image/${mimeType};base64,${imageData.toString('base64')}`;
    
    // Create the image-to-video job
    console.log('Sending request to RunwayML...');
    const imageToVideo = await client.imageToVideo.create({
      model: 'gen3a_turbo',
      promptImage: base64Image,
      promptText: options.description,
    });
    
    console.log(`Job created with ID: ${imageToVideo.id}`);
    
    // Poll for job completion
    let videoData;
    let task;
    
    do {
      // Wait before polling
      console.log('Waiting 10 seconds before checking status...');
      await new Promise(resolve => setTimeout(resolve, 10000));
      
      console.log(`Checking job status...`);
      task = await client.tasks.retrieve(imageToVideo.id);
      console.log(`Current status: ${task.status}`);
      
    } while (!['SUCCEEDED', 'FAILED'].includes(task.status));
    
    if (task.status === 'SUCCEEDED') {
      console.log(`Job completed successfully`);
      
      // Download the video
      let videoUrl;
      
      // Try different ways to access the video URL
      if (task.output) {
        if (Array.isArray(task.output) && task.output.length > 0) {
          // If output is an array, use the first element
          videoUrl = task.output[0];
        } else if (typeof task.output === 'object' && task.output.video) {
          // If output is an object with a video property
          videoUrl = task.output.video;
        }
      } else if (task.result && task.result.videoUrl) {
        videoUrl = task.result.videoUrl;
      } else if (task.videoUrl) {
        videoUrl = task.videoUrl;
      }
      
      if (!videoUrl) {
        console.error('Could not find video URL in task response. Full response:', JSON.stringify(task, null, 2));
        throw new Error('No video URL in completed task');
      }
      
      console.log(`Downloading video from: ${videoUrl}`);
      videoData = await downloadFile(videoUrl);
      
      // Create output directory if it doesn't exist
      const outputDir = path.dirname(options.output);
      await mkdir(outputDir, { recursive: true });
      
      // Save the video
      console.log(`Saving video to: ${options.output}`);
      await writeFile(options.output, videoData);
      
      console.log(`Video generation completed at ${new Date().toLocaleTimeString()}`);
      console.log(`Video saved to: ${options.output}`);
    } else {
      throw new Error(`Job failed: ${task.error || 'Unknown error'}`);
    }
  } catch (error) {
    console.error('Error during video generation:', error);
    process.exit(1);
  }
}

// Run the main function
generateVideo().catch(error => {
  console.error('Unhandled error:', error);
  process.exit(1);
}); 