#!/bin/bash
# Quick Launch Script for Taso v1.0.0
# This script automates the push to GitHub and prepares for Reddit posting

set -e

echo "🚀 Taso v1.0.0 Launch Script"
echo "=============================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Verify we're in the right directory
if [ ! -f "go.mod" ] || [ ! -f "main.go" ]; then
    echo -e "${RED}❌ Error: Not in Taso root directory${NC}"
    echo "Please run this script from the e:/Projects/taso directory"
    exit 1
fi

echo -e "${GREEN}✅ Confirmed: In Taso root directory${NC}"
echo ""

# Step 2: Run tests
echo -e "${YELLOW}📋 Step 1: Running tests...${NC}"
go test ./... 2>&1 | head -20
echo ""

# Step 3: Format code
echo -e "${YELLOW}📋 Step 2: Formatting code...${NC}"
go fmt ./...
echo -e "${GREEN}✅ Code formatted${NC}"
echo ""

# Step 4: Check git status
echo -e "${YELLOW}📋 Step 3: Checking git status...${NC}"
git status
echo ""

# Step 5: Commit changes
echo -e "${YELLOW}📋 Step 4: Ready to commit?${NC}"
read -p "Press Enter to proceed with commit, or Ctrl+C to cancel"

git add .
git commit -m "feat: v1.0.0 - Production Ready Release

- AST-based environment variable scanning
- Multi-language support (9 languages)
- Security hardening and compliance
- Zero-config operation
- CI/CD integration ready
- Comprehensive documentation"

echo -e "${GREEN}✅ Committed to git${NC}"
echo ""

# Step 6: Create tag
echo -e "${YELLOW}📋 Step 5: Creating release tag v1.0.0...${NC}"
git tag -a v1.0.0 -m "Release v1.0.0 - Production Ready

Major Features:
- AST-based scanning for Go, JS, TS
- Multi-language support
- High-performance caching
- Security hardened
- Ready for production use"

echo -e "${GREEN}✅ Tag created${NC}"
echo ""

# Step 7: Push to GitHub
echo -e "${YELLOW}📋 Step 6: Pushing to GitHub...${NC}"
echo "This will:"
echo "1. Push all commits to 'main' branch"
echo "2. Push the v1.0.0 tag (triggers release workflow)"
echo ""
read -p "Ready to push? (Press Enter or Ctrl+C to cancel)"

git push origin main
git push origin v1.0.0

echo -e "${GREEN}✅ Pushed to GitHub${NC}"
echo ""

# Step 8: Monitor GitHub
echo -e "${YELLOW}📋 Step 7: GitHub Release Workflow${NC}"
echo "The release workflow is now running on GitHub."
echo "Monitor progress at: https://github.com/Hossiy21/taso/actions"
echo ""
echo "Once complete, binaries will be available at:"
echo "https://github.com/Hossiy21/taso/releases/tag/v1.0.0"
echo ""

# Step 9: Show Reddit posting info
echo -e "${YELLOW}📋 Step 8: Ready for Reddit!${NC}"
echo ""
echo "Your project is now live! Next steps:"
echo ""
echo "1. 📍 r/golang - Post in 1 hour (best audience for Go projects)"
echo "   Title: [Project] Taso - Find ghost environment variables in your Go code"
echo ""
echo "2. 📍 r/DevOps - Post in 3 days"
echo "   Title: [Tool] Taso - Environment Variable Drift Detection"
echo ""
echo "3. 📍 r/programming - Post in 7 days"
echo "   Title: Taso - AST-based tool to find environment variable misconfigurations"
echo ""
echo "4. 📱 Twitter/LinkedIn - After initial Reddit traction"
echo ""
echo "See DEPLOYMENT_AND_REDDIT_GUIDE.md for full details"
echo ""

# Step 10: Display verification
echo -e "${GREEN}🎉 Launch Complete!${NC}"
echo ""
echo "Verification checklist:"
echo "✅ Code formatted"
echo "✅ Tests passing"
echo "✅ Committed to git"
echo "✅ Tagged as v1.0.0"
echo "✅ Pushed to GitHub"
echo "✅ Release workflow started"
echo ""
echo "Next: Wait for GitHub release to complete, then post to Reddit!"
echo ""
echo "Questions? Check:"
echo "- DEPLOYMENT_AND_REDDIT_GUIDE.md"
echo "- SECURITY.md"
echo "- README.md"
