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
        local additions=$(git diff --numstat HEAD "$file" 2>/dev/null | cut -f1)
        local deletions=$(git diff --numstat HEAD "$file" 2>/dev/null | cut -f2)
        if [[ ! -z "$additions" && ! -z "$deletions" ]]; then
            echo -e "    📈 Changes: ${GREEN}+$additions${NC} / ${RED}-$deletions${NC} lines"
        fi
        
        # Show git diff with some context
        if git ls-files --error-unmatch "$file" >/dev/null 2>&1; then
            echo -e "    📝 Diff:"
            git diff --color=always HEAD "$file" | sed 's/^/      /'
        else
            echo -e "    📝 New file content:"
            head -n 10 "$file" | sed 's/^/      /'
            local total_lines=$(wc -l < "$file")
            if [ "$total_lines" -gt 10 ]; then
                echo "      ... (showing first 10 of $total_lines lines)"
            fi
        fi
    else
        echo -e "    ${RED}File has been deleted${NC}"
    fi
}

echo -e "${YELLOW}📋 Changes Since Last Commit${NC}"

# Get current branch name and last commit info
BRANCH=$(git rev-parse --abbrev-ref HEAD)
LAST_COMMIT=$(git log -1 --pretty=format:"%h - %s (%cr)" 2>/dev/null)
if [ ! -z "$LAST_COMMIT" ]; then
    echo -e "\n${BLUE}🔖 Current Branch: ${CYAN}$BRANCH${NC}"
    echo -e "${BLUE}📅 Last Commit: ${CYAN}$LAST_COMMIT${NC}"
fi

# Check for uncommitted changes
if [[ $(git status --porcelain) ]]; then
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
else
    echo -e "\n${GREEN}✨ No uncommitted changes${NC}"
fi

# Show unpushed commits if any
echo -e "\n${YELLOW}📤 Unpushed Commits:${NC}"
UNPUSHED=$(git log @{u}.. --oneline 2>/dev/null)
if [ ! -z "$UNPUSHED" ]; then
    echo -e "${BLUE}The following commits haven't been pushed to GitHub:${NC}"
    echo "$UNPUSHED" | sed 's/^/  /'
else
    echo -e "${GREEN}No unpushed commits${NC}"
fi 