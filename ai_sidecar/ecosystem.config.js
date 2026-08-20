// PM2 Process Manager Configuration for freel-ai-sidecar
//
// Simple meaning:
//   This file tells PM2 how to run, monitor, and scale the Python FastAPI service.
//   It ensures that uvicorn runs within the project's virtual environment python context,
//   and automatically restarts the app if it crashes or exceeds memory limits.
//
// To start:
//   pm2 start ecosystem.config.js
//
// To reload:
//   pm2 reload freel-ai-sidecar

//If your friend asks:

// "What are these two files?"

// Say:

// ecosystem.config.js tells PM2 how to run and monitor our Python AI service. deploy.sh automates the deployment by creating the Python environment, installing dependencies, and starting or reloading the AI service through PM2 on EC2.

// Or even simpler:

// ecosystem.config.js
// = HOW to run the AI service

// deploy.sh
// = HOW to deploy/update the AI service

// That's the main thing to remember.

module.exports = {
  apps: [
    {
      name: "freel-ai-sidecar",
      script: "./venv/bin/uvicorn",
      args: "main:app --host 0.0.0.0 --port 8090",
      interpreter: "none", // Skip PM2 node/python default wrappers and execute raw uvicorn directly
      cwd: "/Users/varun.kanade/go/src/freel/freel-project/ai_sidecar",
      instances: 1,
      autorestart: true,
      watch: false, // Set to false in production to prevent recursive reload loops on logs/database updates
      max_memory_restart: "400M", // Throttles memory footprint to safeguard standard micro EC2 instances
      env: {
        NODE_ENV: "production",
        PYTHONUNBUFFERED: "1",
        PORT: "8090"
      }
    }
  ]
};
