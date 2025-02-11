#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m' # No Color

echo -e "${YELLOW}🔄 Syncing with GitHub...${NC}"

# Get current branch name
BRANCH=$(git rev-parse --abbrev-ref HEAD)

# Fetch latest changes from remote
echo -e "\n${YELLOW}📥 Fetching latest changes...${NC}"
git fetch origin

# Function to show detailed changes for a file
show_file_changes() {
    local file=$1
    echo -e "\n${BLUE}📄 Changes in ${BOLD}$file${NC}:"
    if [[ -f "$file" ]]; then
        # Show git diff with some context
        git diff --color=always HEAD "$file" | sed 's/^/    /'
    else
        echo -e "    ${RED}File has been deleted${NC}"
    fi
}

# Check if there are uncommitted changes
if [[ $(git status --porcelain) ]]; then
    # Show status summary
    echo -e "\n${YELLOW}📝 Changed files:${NC}"
    git status --short

    # Show detailed changelog
    echo -e "\n${YELLOW}📋 Detailed Changelog:${NC}"
    
    # Show new files
    echo -e "\n${GREEN}✨ New Files:${NC}"
    git status --porcelain | grep "^??" | cut -d' ' -f2- | while read -r file; do
        echo -e "  + ${file}"
    done

    # Show modified files with diff
    echo -e "\n${YELLOW}📝 Modified Files:${NC}"
    git status --porcelain | grep "^.M" | cut -d' ' -f2- | while read -r file; do
        show_file_changes "$file"
    done

    # Show deleted files
    echo -e "\n${RED}🗑️  Deleted Files:${NC}"
    git status --porcelain | grep "^.D" | cut -d' ' -f2- | while read -r file; do
        echo -e "  - ${file}"
    done

    # Show rename operations
    echo -e "\n${BLUE}📋 Renamed Files:${NC}"
    git status --porcelain | grep "^R" | while read -r line; do
        old_file=$(echo "$line" | cut -d' ' -f2-)
        new_file=$(echo "$line" | cut -d' ' -f3-)
        echo -e "  ${old_file} → ${new_file}"
    done

    # Summary of changes
    echo -e "\n${YELLOW}📊 Changes Summary:${NC}"
    echo -e "  Modified: $(git status --porcelain | grep "^.M" | wc -l | tr -d ' ') files"
    echo -e "  Added:    $(git status --porcelain | grep "^??" | wc -l | tr -d ' ') files"
    echo -e "  Deleted:  $(git status --porcelain | grep "^.D" | wc -l | tr -d ' ') files"
    echo -e "  Renamed:  $(git status --porcelain | grep "^R" | wc -l | tr -d ' ') files"

    # Get commit message
    echo -e "\n${YELLOW}💭 Enter commit message (or press enter for default):${NC}"
    read -r commit_msg
    if [ -z "$commit_msg" ]; then
        commit_msg="Update: $(date +'%Y-%m-%d %H:%M:%S')"
    fi

    # Add all changes
    echo -e "\n${YELLOW}➕ Adding all changes...${NC}"
    git add .

    # Commit changes
    echo -e "\n${YELLOW}📦 Committing changes...${NC}"
    git commit -m "$commit_msg"
fi

# Show incoming changes if any
echo -e "\n${YELLOW}⬇️ Checking for incoming changes...${NC}"
INCOMING_CHANGES=$(git log HEAD..origin/$BRANCH --oneline)
if [ ! -z "$INCOMING_CHANGES" ]; then
    echo -e "${BLUE}Incoming changes from GitHub:${NC}"
    echo "$INCOMING_CHANGES" | sed 's/^/  /'
fi

# Pull latest changes
echo -e "\n${YELLOW}⬇️ Pulling latest changes...${NC}"
git pull origin $BRANCH

# Push changes
echo -e "\n${YELLOW}⬆️ Pushing changes...${NC}"
git push origin $BRANCH

echo -e "\n${GREEN}✅ Sync complete!${NC}" 