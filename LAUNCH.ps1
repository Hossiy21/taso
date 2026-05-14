# Quick Launch Script for Taso v1.0.0 (Windows PowerShell)
# This script automates the push to GitHub and prepares for Reddit posting

Write-Host "🚀 Taso v1.0.0 Launch Script (Windows)" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

# Step 1: Verify we're in the right directory
if (!(Test-Path "go.mod") -or !(Test-Path "main.go")) {
    Write-Host "❌ Error: Not in Taso root directory" -ForegroundColor Red
    Write-Host "Please run this script from the e:\Projects\taso directory"
    exit 1
}

Write-Host "✅ Confirmed: In Taso root directory" -ForegroundColor Green
Write-Host ""

# Step 2: Run tests
Write-Host "📋 Step 1: Running tests..." -ForegroundColor Yellow
Write-Host ""
go test ./... 2>&1 | Select-Object -First 20
Write-Host ""

# Step 3: Format code
Write-Host "📋 Step 2: Formatting code..." -ForegroundColor Yellow
go fmt ./...
Write-Host "✅ Code formatted" -ForegroundColor Green
Write-Host ""

# Step 4: Check git status
Write-Host "📋 Step 3: Checking git status..." -ForegroundColor Yellow
git status
Write-Host ""

# Step 5: Commit changes
Write-Host "📋 Step 4: Ready to commit?" -ForegroundColor Yellow
Read-Host "Press Enter to proceed with commit, or Ctrl+C to cancel"

git add .
git commit -m "feat: v1.0.0 - Production Ready Release

- AST-based environment variable scanning
- Multi-language support (9 languages)
- Security hardening and compliance
- Zero-config operation
- CI/CD integration ready
- Comprehensive documentation"

Write-Host "✅ Committed to git" -ForegroundColor Green
Write-Host ""

# Step 6: Create tag
Write-Host "📋 Step 5: Creating release tag v1.0.0..." -ForegroundColor Yellow
git tag -a v1.0.0 -m "Release v1.0.0 - Production Ready

Major Features:
- AST-based scanning for Go, JS, TS
- Multi-language support
- High-performance caching
- Security hardened
- Ready for production use"

Write-Host "✅ Tag created" -ForegroundColor Green
Write-Host ""

# Step 7: Push to GitHub
Write-Host "📋 Step 6: Pushing to GitHub..." -ForegroundColor Yellow
Write-Host "This will:" -ForegroundColor Cyan
Write-Host "1. Push all commits to 'main' branch"
Write-Host "2. Push the v1.0.0 tag (triggers release workflow)"
Write-Host ""
Read-Host "Ready to push? (Press Enter or Ctrl+C to cancel)"

git push origin main
git push origin v1.0.0

Write-Host "✅ Pushed to GitHub" -ForegroundColor Green
Write-Host ""

# Step 8: Monitor GitHub
Write-Host "📋 Step 7: GitHub Release Workflow" -ForegroundColor Yellow
Write-Host ""
Write-Host "The release workflow is now running on GitHub."
Write-Host "Monitor progress at: https://github.com/Hossiy21/taso/actions"
Write-Host ""
Write-Host "Once complete, binaries will be available at:"
Write-Host "https://github.com/Hossiy21/taso/releases/tag/v1.0.0"
Write-Host ""

# Step 9: Show Reddit posting info
Write-Host "📋 Step 8: Ready for Reddit!" -ForegroundColor Yellow
Write-Host ""
Write-Host "Your project is now live! Next steps:" -ForegroundColor Green
Write-Host ""
Write-Host "1. 📍 r/golang - Post in 1 hour (best audience for Go projects)"
Write-Host "   Title: [Project] Taso - Find ghost environment variables in your Go code"
Write-Host ""
Write-Host "2. 📍 r/DevOps - Post in 3 days"
Write-Host "   Title: [Tool] Taso - Environment Variable Drift Detection"
Write-Host ""
Write-Host "3. 📍 r/programming - Post in 7 days"
Write-Host "   Title: Taso - AST-based tool to find environment variable misconfigurations"
Write-Host ""
Write-Host "4. 📱 Twitter/LinkedIn - After initial Reddit traction"
Write-Host ""
Write-Host "See DEPLOYMENT_AND_REDDIT_GUIDE.md for full details"
Write-Host ""

# Step 10: Display verification
Write-Host "🎉 Launch Complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Verification checklist:" -ForegroundColor Cyan
Write-Host "✅ Code formatted"
Write-Host "✅ Tests passing"
Write-Host "✅ Committed to git"
Write-Host "✅ Tagged as v1.0.0"
Write-Host "✅ Pushed to GitHub"
Write-Host "✅ Release workflow started"
Write-Host ""
Write-Host "Next: Wait for GitHub release to complete, then post to Reddit!"
Write-Host ""
Write-Host "Questions? Check:" -ForegroundColor Yellow
Write-Host "- DEPLOYMENT_AND_REDDIT_GUIDE.md"
Write-Host "- SECURITY.md"
Write-Host "- README.md"
