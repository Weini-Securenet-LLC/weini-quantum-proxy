#!/bin/bash

# Weini Securenet LLC - Publication Script
# Weini Quantum Proxy v0.9.0-beta

set -e  # Exit on error

echo "🚀 Weini Quantum Proxy - Publication Script"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📦 Project: Weini Quantum Proxy v0.9.0-beta"
echo "🏢 Organization: Weini Securenet LLC"
echo "🌐 Website: https://weinidaohang.com/"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check if we're in the right directory
if [ ! -f "README.md" ] || [ ! -f "LICENSE" ]; then
    echo "❌ Error: Please run this script from the project root directory"
    exit 1
fi

echo "✅ Project directory verified"
echo ""

# Step 1: Initialize Git
echo "📝 Step 1: Initializing Git repository..."
if [ ! -d ".git" ]; then
    git init
    echo "✅ Git repository initialized"
else
    echo "ℹ️  Git repository already exists"
fi
echo ""

# Step 2: Configure Git
echo "📝 Step 2: Configuring Git..."
git config user.name "Weini Securenet"
git config user.email "weinidaohang@proton.me"
echo "✅ Git configured"
echo ""

# Step 3: Add files
echo "📝 Step 3: Adding files to Git..."
git add .
echo "✅ Files added"
echo ""

# Step 4: Create initial commit
echo "📝 Step 4: Creating initial commit..."
git commit -m "feat: initial beta release v0.9.0

Weini Quantum Proxy - First Public Beta

Powered by Weini Securenet LLC
Digital Human Rights & Security Ecosystem

Core Features:
- Multi-protocol support (SS/VMess/VLESS/Trojan)
- Cross-platform desktop application (Windows/macOS/Linux)
- Intelligent node discovery and validation
- AI Agent Skills for automation
- Mobile device configuration guides
- Multi-language support (EN/ZH/FA/AR/RU)

Website: https://weinidaohang.com/
Contact: weinidaohang@proton.me
GitHub: https://github.com/Weini-Securenet-LLC

This is a beta release for community testing and feedback."

echo "✅ Initial commit created"
echo ""

# Step 5: Add remote
echo "📝 Step 5: Adding GitHub remote..."
REMOTE_URL="git@github.com:Weini-Securenet-LLC/weini-quantum-proxy.git"

if git remote | grep -q "origin"; then
    echo "ℹ️  Remote 'origin' already exists, updating..."
    git remote set-url origin $REMOTE_URL
else
    git remote add origin $REMOTE_URL
fi

echo "✅ Remote added: $REMOTE_URL"
echo ""

# Step 6: Rename branch to main
echo "📝 Step 6: Renaming branch to main..."
git branch -M main
echo "✅ Branch renamed to main"
echo ""

# Step 7: Ask for confirmation before pushing
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "⚠️  CONFIRMATION REQUIRED"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "You are about to push to:"
echo "  Repository: Weini-Securenet-LLC/weini-quantum-proxy"
echo "  Branch: main"
echo "  Commit: Initial beta release v0.9.0"
echo ""
read -p "Continue with push? (y/N): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo "❌ Publication cancelled by user"
    echo ""
    echo "ℹ️  To publish later, run:"
    echo "   git push -u origin main"
    echo "   git tag -a v0.9.0-beta -m 'Beta release'"
    echo "   git push origin v0.9.0-beta"
    exit 0
fi

echo ""

# Step 8: Push to GitHub
echo "📝 Step 7: Pushing to GitHub..."
git push -u origin main
echo "✅ Pushed to GitHub"
echo ""

# Step 9: Create and push tag
echo "📝 Step 8: Creating beta tag..."
git tag -a v0.9.0-beta -m "v0.9.0-beta: First Public Beta Release

Weini Quantum Proxy Beta

This is the first public test release of Weini Quantum Proxy,
developed by Weini Securenet LLC.

What's New:
- Complete desktop application for Windows/macOS/Linux
- Multi-language support (EN/ZH/FA/AR/RU)
- AI Agent Skills for node automation
- Comprehensive mobile device guides
- Full documentation and community resources

Known Issues:
- Some translations need refinement
- Mobile native apps not yet available
- Performance optimization ongoing

Beta Test Period: Q2 2026
Stable Release Target: Q3 2026

For support and feedback:
- GitHub: https://github.com/Weini-Securenet-LLC/weini-quantum-proxy
- Website: https://weinidaohang.com/
- Email: weinidaohang@proton.me
- Telegram: https://t.me/weini_quantum

Thank you for being part of our beta program!

---
Powered by Weini Securenet LLC
Digital Human Rights & Security Ecosystem"

git push origin v0.9.0-beta
echo "✅ Beta tag created and pushed"
echo ""

# Success message
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🎉 SUCCESS! Project Published!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "✅ Repository: https://github.com/Weini-Securenet-LLC/weini-quantum-proxy"
echo "✅ Version: v0.9.0-beta"
echo "✅ Status: Pre-release (Beta)"
echo ""
echo "📝 Next Steps:"
echo ""
echo "1. Create GitHub Release:"
echo "   → Visit: https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/releases/new"
echo "   → Tag: v0.9.0-beta"
echo "   → Title: Weini Quantum Proxy v0.9.0-beta - First Public Beta"
echo "   → Copy release notes from: RELEASE_NOTES.md"
echo "   → Check 'This is a pre-release'"
echo "   → Publish release"
echo ""
echo "2. Configure Repository Settings:"
echo "   → Add topics: proxy, vpn, freedom, censorship, ai-tools"
echo "   → Enable Discussions"
echo "   → Enable Issues"
echo "   → Add website: https://weinidaohang.com/"
echo ""
echo "3. Announce Release:"
echo "   → GitHub Discussions"
echo "   → Telegram group"
echo "   → Update company website"
echo ""
echo "4. Monitor Feedback:"
echo "   → Watch GitHub Issues"
echo "   → Respond to discussions"
echo "   → Plan improvements"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🌐 For a freer internet!"
echo "Powered by Weini Securenet LLC"
echo ""
