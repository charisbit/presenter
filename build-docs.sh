#!/bin/bash

# Build Documentation Script for Intelligent Presenter
# Combines Nikola + TypeDoc + Go documentation

set -e

echo "🚀 Building Intelligent Presenter Documentation..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if we're in the project root
if [ ! -f "go.work" ] && [ ! -f "package.json" ] && [ ! -d "backend" ]; then
    echo -e "${RED}❌ Error: Please run this script from the project root directory${NC}"
    exit 1
fi

# Function to print colored output
print_step() {
    echo -e "${BLUE}📝 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Create docs directory if it doesn't exist
mkdir -p docs-site/files

print_step "Step 1: Generating Go backend documentation..."

# Generate Go documentation
cd backend
if command -v go &> /dev/null; then
    echo "Module information:" > ../docs-site/files/go-docs.txt
    go list -m >> ../docs-site/files/go-docs.txt
    echo "" >> ../docs-site/files/go-docs.txt
    echo "=== Package Documentation ===" >> ../docs-site/files/go-docs.txt
    echo "" >> ../docs-site/files/go-docs.txt
    
    # Find all Go files and extract documentation
    find . -name "*.go" -not -path "./vendor/*" | while read -r file; do
        echo "=== $file ===" >> ../docs-site/files/go-docs.txt
        head -50 "$file" >> ../docs-site/files/go-docs.txt
        echo "" >> ../docs-site/files/go-docs.txt
    done
    
    print_success "Go documentation generated"
else
    print_warning "Go not found, skipping Go documentation"
fi
cd ..

print_step "Step 2: Generating TypeScript frontend documentation..."

# Generate TypeScript documentation with TypeDoc
cd frontend
if command -v npm &> /dev/null && [ -f "package.json" ]; then
    # Check if TypeDoc is available
    if npm list typedoc &> /dev/null || npm list -g typedoc &> /dev/null; then
        npx typedoc \
            --entryPointStrategy expand \
            --out ../docs-site/files/typescript-docs \
            src \
            --excludeExternals \
            --exclude "**/*.test.ts" \
            --exclude "**/test-*.ts" \
            --skipErrorChecking \
            --name "Intelligent Presenter Frontend" \
            --readme none
        print_success "TypeScript documentation generated"
    else
        print_warning "TypeDoc not found, installing..."
        npm install typedoc --save-dev
        npx typedoc \
            --entryPointStrategy expand \
            --out ../docs-site/files/typescript-docs \
            src \
            --excludeExternals \
            --exclude "**/*.test.ts" \
            --exclude "**/test-*.ts" \
            --skipErrorChecking \
            --name "Intelligent Presenter Frontend" \
            --readme none
        print_success "TypeScript documentation generated"
    fi
else
    print_warning "npm or package.json not found, skipping TypeScript documentation"
fi
cd ..

print_step "Step 3: Building Nikola documentation site..."

# Build Nikola site
cd docs-site
if command -v nikola &> /dev/null; then
    # Clean previous build
    if [ -d "output" ]; then
        rm -rf output/*
    fi
    
    # Build the site
    nikola build
    print_success "Nikola site built successfully"
    
    # Check if build was successful
    if [ -f "output/index.html" ]; then
        print_success "Documentation site ready at docs-site/output/"
    else
        print_error "Nikola build failed - no index.html found"
        exit 1
    fi
else
    print_error "Nikola not found. Please install: pip install 'nikola[extras]'"
    exit 1
fi
cd ..

print_step "Step 4: Creating documentation archive..."

# Create a timestamp for the archive
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
ARCHIVE_NAME="intelligent-presenter-docs-$TIMESTAMP.tar.gz"

# Create archive of the documentation
tar -czf "$ARCHIVE_NAME" \
    --exclude="*.pyc" \
    --exclude="__pycache__" \
    docs-site/output/ \
    docs-site/files/go-docs.txt \
    docs-site/files/typescript-docs/

print_success "Documentation archive created: $ARCHIVE_NAME"

print_success "Documentation build completed!"

echo ""
echo "📊 Build Summary:"
echo "===================="
echo "📁 Output directory: docs-site/output/"
echo "📦 Archive created: $ARCHIVE_NAME"
echo ""
echo "🌐 To view the documentation:"
echo "   cd docs-site/output && python -m http.server 8000"
echo "   Then open: http://localhost:8000"
echo ""

# Optional: serve the documentation locally
if [[ "$1" == "--serve" ]]; then
    print_step "Starting local documentation server..."
    cd docs-site
    nikola serve &
    echo "Documentation available at: http://localhost:8000"
    echo "Press Ctrl+C to stop the server"
    wait
fi

echo "🚀 Documentation is ready for deployment!"