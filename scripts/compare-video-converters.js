#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');
const { program } = require('commander');

// Parse command line arguments
program
  .option('-i, --image <path>', 'Path to a single image to convert', 'output/images/image_bridgerton_vs._stric_09fba2da.png')
  .option('-o, --output-dir <dir>', 'Directory for output videos', 'output/videos/comparison')
  .parse(process.argv);

const options = program.opts();

// Ensure output directory exists
if (!fs.existsSync(options.outputDir)) {
  fs.mkdirSync(options.outputDir, { recursive: true });
}

console.log('='.repeat(80));
console.log('COMPARING VIDEO CONVERTERS');
console.log('='.repeat(80));
console.log(`Input image: ${options.image}`);
console.log(`Output directory: ${options.outputDir}`);
console.log('='.repeat(80));

// Function to run a command and return its output
function runCommand(command) {
  console.log(`Running: ${command}`);
  try {
    const output = execSync(command, { encoding: 'utf8' });
    return { success: true, output };
  } catch (error) {
    return { 
      success: false, 
      output: error.stdout || '', 
      error: error.stderr || error.message 
    };
  }
}

// Run the RunwayML converter
console.log('\n1. RUNNING RUNWAYML CONVERTER');
console.log('-'.repeat(80));
const runwayResult = runCommand(`node scripts/video-converter.js --input-dir "${options.image}" --output-dir "${options.outputDir}/runway"`);

if (runwayResult.success) {
  console.log('RunwayML converter completed successfully');
} else {
  console.error('RunwayML converter failed:');
  console.error(runwayResult.error);
}

// Run the Replicate converter
console.log('\n2. RUNNING REPLICATE CONVERTER');
console.log('-'.repeat(80));
const replicateResult = runCommand(`node scripts/video-converter-replicate.js --input-dir "${options.image}" --output-dir "${options.outputDir}/replicate"`);

if (replicateResult.success) {
  console.log('Replicate converter completed successfully');
} else {
  console.error('Replicate converter failed:');
  console.error(replicateResult.error);
}

// Summary
console.log('\n='.repeat(80));
console.log('COMPARISON SUMMARY');
console.log('='.repeat(80));

// Find the generated videos
const runwayVideos = fs.existsSync(`${options.outputDir}/runway`) 
  ? fs.readdirSync(`${options.outputDir}/runway`).filter(f => f.endsWith('.mp4'))
  : [];

const replicateVideos = fs.existsSync(`${options.outputDir}/replicate`) 
  ? fs.readdirSync(`${options.outputDir}/replicate`).filter(f => f.endsWith('.mp4'))
  : [];

console.log(`RunwayML videos: ${runwayVideos.length > 0 ? runwayVideos.join(', ') : 'None'}`);
console.log(`Replicate videos: ${replicateVideos.length > 0 ? replicateVideos.join(', ') : 'None'}`);

console.log('\nTo view the videos, open:');
if (runwayVideos.length > 0) {
  console.log(`- ${path.resolve(options.outputDir, 'runway', runwayVideos[0])}`);
}
if (replicateVideos.length > 0) {
  console.log(`- ${path.resolve(options.outputDir, 'replicate', replicateVideos[0])}`);
}

console.log('\nComparison complete!'); 