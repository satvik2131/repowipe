#!/bin/bash
# deploy.sh - Place this in your repowipe/ root directory

set -e  # Exit on any error

echo "🏗️  Building React frontend..."
cd frontend
npm ci  # Use ci for faster, reliable builds
npm run build
cd ..

echo "🔨 Building Go backend binary for Linux..."
cd backend

# Build optimized binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o repowipe .

echo "📋 Checking build artifacts..."
echo "Binary size: $(du -h repowipe | cut -f1)"
echo "Static files: $(du -sh static | cut -f1)"

# Create railway.toml if it doesn't exist
if [ ! -f railway.toml ]; then
    echo "📝 Creating railway.toml..."
    cat > railway.toml << 'EOF'
[build]
buildCommand = "echo 'Using pre-built binary'"

[deploy]
startCommand = "./repowipe"
healthcheckPath = "/health"
healthcheckTimeout = 100
restartPolicyType = "on_failure"

[environments.production.variables]
GIN_MODE = "release"
EOF
fi

echo "🚀 Setting up Railway project..."
# Check if already linked to a project
if railway status &>/dev/null; then
    echo "Already linked to Railway project"
else
    echo "Creating new Railway project..."
    railway init repowipe-app
fi

echo "🚀 Deploying to Railway..."
railway up

echo "✅ Deployment complete!"
echo ""
echo "📊 Useful commands:"
echo "  railway status  - Check deployment status"
echo "  railway logs    - View application logs"
echo "  railway open    - Open in browser"
echo "  railway domain  - Manage custom domains"