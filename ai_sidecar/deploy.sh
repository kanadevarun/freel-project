#!/bin/bash
# freel-ai-sidecar EC2 deployment script
#
# Simple meaning:
#   This script automates the setup, virtual env creation, pip installation,
#   and PM2 process hot-reload for the Python AI Sidecar on standard AWS EC2 instances.

set -e

# Get script's parent directory dynamically to enable portable execution on AWS EC2
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== starting deployment in $SCRIPT_DIR ==="

# 1. Setup Virtual Environment if it does not exist
if [ ! -d "venv" ]; then
    echo "Creating virtual environment..."
    python3 -m venv venv
fi

# 2. Activate virtual environment and upgrade pip
source venv/bin/activate
echo "Upgrading pip..."
pip install --upgrade pip

# 3. Install requirements
echo "Installing pip requirements..."
pip install -r requirements.txt

# 4. Check if PM2 is installed globally
if ! command -v pm2 &> /dev/null; then
    echo "PM2 is not installed. Please install node and PM2 globally (e.g. npm install -g pm2)."
    exit 1
fi

# 5. Start or reload sidecar in PM2
echo "Deploying to PM2..."
if pm2 list | grep -q "freel-ai-sidecar"; then
    echo "freel-ai-sidecar service is running, reloading..."
    pm2 reload ecosystem.config.js
else
    echo "freel-ai-sidecar service not registered, starting new instance..."
    pm2 start ecosystem.config.js
fi

# 6. Save PM2 configuration to persist across system reboots
pm2 save

echo "=== deployment completed successfully ==="
