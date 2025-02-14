#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
BOLD='\033[1m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to calculate lines of code
count_lines() {
    local file=$1
    if [[ -f "$file" ]]; then
        wc -l < "$file" | tr -d ' '
    else
        echo "0"
    fi
}

# Function to get file size in human-readable format
get_file_size() {
    local file=$1
    if [[ -f "$file" ]]; then
        if [[ "$OSTYPE" == "darwin"* ]]; then
            stat -f %z "$file" | awk '{ split( "B KB MB GB TB" , v ); s=1; while( $1>1024 ){ $1/=1024; s++ } printf "%.1f %s", $1, v[s] }'
        else
            stat -f %s "$file" | awk '{ split( "B KB MB GB TB" , v ); s=1; while( $1>1024 ){ $1/=1024; s++ } printf "%.1f %s", $1, v[s] }'
        fi
    else
        echo "0B"
    fi
}

# Function to show detailed changes for a file
show_file_changes() {
    local file=$1
    echo -e "\n${BLUE}📄 Changes in ${BOLD}$file${NC}:"
    if [[ -f "$file" ]]; then
        # Show file stats
        local lines=$(count_lines "$file")
        local size=$(get_file_size "$file")
        local ext="${file##*.}"
        echo -e "    📊 Stats: ${CYAN}$lines lines${NC}, ${CYAN}$size${NC}, Type: ${CYAN}.$ext${NC}"
        
        # Count changes
        local additions=$(git diff --numstat HEAD "$file" | cut -f1)
        local deletions=$(git diff --numstat HEAD "$file" | cut -f2)
        if [[ ! -z "$additions" && ! -z "$deletions" ]]; then
            echo -e "    📈 Changes: ${GREEN}+$additions${NC} / ${RED}-$deletions${NC} lines"
        fi
        
        # Show git diff with some context
        echo -e "    📝 Diff:"
        git diff --color=always HEAD "$file" | sed 's/^/      /'
    else
        echo -e "    ${RED}File has been deleted${NC}"
    fi
}

echo -e "${YELLOW}🔄 Syncing with GitHub...${NC}"

# Get current branch name
BRANCH=$(git rev-parse --abbrev-ref HEAD)

# Fetch latest changes from remote
echo -e "\n${YELLOW}📥 Fetching latest changes...${NC}"
git fetch origin

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
        size=$(get_file_size "$file")
        lines=$(count_lines "$file")
        echo -e "  + ${file} (${CYAN}$lines lines${NC}, ${CYAN}$size${NC})"
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
        size=$(get_file_size "$new_file")
        lines=$(count_lines "$new_file")
        echo -e "  ${old_file} → ${new_file} (${CYAN}$lines lines${NC}, ${CYAN}$size${NC})"
    done

    # Calculate total changes
    total_files=$(git status --porcelain | wc -l | tr -d ' ')
    total_additions=$(git diff --numstat | awk '{sum += $1} END {print sum}')
    total_deletions=$(git diff --numstat | awk '{sum += $2} END {print sum}')

    # Summary of changes
    echo -e "\n${YELLOW}📊 Changes Summary:${NC}"
    echo -e "  Modified: $(git status --porcelain | grep "^.M" | wc -l | tr -d ' ') files"
    echo -e "  Added:    $(git status --porcelain | grep "^??" | wc -l | tr -d ' ') files"
    echo -e "  Deleted:  $(git status --porcelain | grep "^.D" | wc -l | tr -d ' ') files"
    echo -e "  Renamed:  $(git status --porcelain | grep "^R" | wc -l | tr -d ' ') files"
    echo -e "  ${GREEN}Total: $total_files files changed, +$total_additions insertions, -$total_deletions deletions${NC}"

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
    
    # Show detailed incoming changes
    echo -e "\n${BLUE}📋 Detailed incoming changes:${NC}"
    git log HEAD..origin/$BRANCH --stat --color | sed 's/^/  /'
fi

# Pull latest changes
echo -e "\n${YELLOW}⬇️ Pulling latest changes...${NC}"
git pull origin $BRANCH

# Push changes
echo -e "\n${YELLOW}⬆️ Pushing changes...${NC}"
git push origin $BRANCH

echo -e "\n${GREEN}✅ Sync complete!${NC}" 