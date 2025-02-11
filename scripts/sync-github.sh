#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}🔄 Syncing with GitHub...${NC}"

# Get current branch name
BRANCH=$(git rev-parse --abbrev-ref HEAD)

# Fetch latest changes from remote
echo -e "\n${YELLOW}📥 Fetching latest changes...${NC}"
git fetch origin

# Check if there are uncommitted changes
if [[ $(git status --porcelain) ]]; then
    # Show status
    echo -e "\n${YELLOW}📝 Found uncommitted changes:${NC}"
    git status --short

    # Add all changes
    echo -e "\n${YELLOW}➕ Adding all changes...${NC}"
    git add .

    # Get commit message
    echo -e "\n${YELLOW}💭 Enter commit message (or press enter for default):${NC}"
    read -r commit_msg
    if [ -z "$commit_msg" ]; then
        commit_msg="Update: $(date +'%Y-%m-%d %H:%M:%S')"
    fi

    # Commit changes
    echo -e "\n${YELLOW}📦 Committing changes...${NC}"
    git commit -m "$commit_msg"
fi

# Pull latest changes
echo -e "\n${YELLOW}⬇️ Pulling latest changes...${NC}"
git pull origin $BRANCH

# Push changes
echo -e "\n${YELLOW}⬆️ Pushing changes...${NC}"
git push origin $BRANCH

echo -e "\n${GREEN}✅ Sync complete!${NC}" 