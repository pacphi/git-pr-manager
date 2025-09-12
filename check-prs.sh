#!/bin/bash

# check-prs.sh - Check PR status across multiple repositories
# This script reads the config.yaml file and checks for open PRs that are ready to merge

set -euo pipefail

# Configuration
CONFIG_FILE="${CONFIG_FILE:-config.yaml}"
OUTPUT_FORMAT="${OUTPUT_FORMAT:-table}"  # table, json, csv

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Log functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" >&2
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" >&2
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" >&2
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Detect operating system
detect_os() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "linux"
    else
        echo "unknown"
    fi
}

# Get platform-specific install command
get_install_command() {
    local os
    os=$(detect_os)

    case "$os" in
        "macos")
            echo "brew install yq jq"
            ;;
        "linux")
            if command -v apt-get &> /dev/null; then
                echo "sudo apt-get update && sudo apt-get install jq && sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 && sudo chmod +x /usr/local/bin/yq"
            elif command -v yum &> /dev/null; then
                echo "sudo yum install jq && sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 && sudo chmod +x /usr/local/bin/yq"
            elif command -v dnf &> /dev/null; then
                echo "sudo dnf install jq && sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 && sudo chmod +x /usr/local/bin/yq"
            elif command -v pacman &> /dev/null; then
                echo "sudo pacman -S jq yq"
            else
                echo "Install jq and yq using your package manager, or see: https://github.com/mikefarah/yq#install"
            fi
            ;;
        *)
            echo "Install jq and yq using your package manager"
            ;;
    esac
}

# Check dependencies
check_dependencies() {
    local missing_deps=()

    if ! command -v yq &> /dev/null; then
        missing_deps+=("yq")
    fi

    if ! command -v jq &> /dev/null; then
        missing_deps+=("jq")
    fi

    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        log_info "Install with: $(get_install_command)"
        exit 1
    fi
}

# Parse YAML configuration
parse_config() {
    if [[ ! -f "$CONFIG_FILE" ]]; then
        log_error "Configuration file not found: $CONFIG_FILE"
        exit 1
    fi
}

# Check GitHub PRs
check_github_prs() {
    local repo_name="$1"
    local repo_url="$2"
    local allowed_actors
    local skip_labels

    # Get configuration
    allowed_actors=$(yq '.config.pr_filters.allowed_actors[]' "$CONFIG_FILE" 2>/dev/null || echo "")
    skip_labels=$(yq '.config.pr_filters.skip_labels[]' "$CONFIG_FILE" 2>/dev/null || echo "")

    if [[ -z "$GITHUB_TOKEN" ]]; then
        log_error "GITHUB_TOKEN environment variable not set"
        return 1
    fi

    log_info "Checking GitHub repository: $repo_name"

    # Get open PRs
    local prs_json
    prs_json=$(gh api "repos/$repo_name/pulls" --jq '.[] | select(.state == "open")' 2>/dev/null || echo "[]")

    if [[ "$prs_json" == "[]" ]] || [[ -z "$prs_json" ]]; then
        echo "No open PRs found for $repo_name"
        return 0
    fi

    # Process each PR
    echo "$prs_json" | jq -r '.' | while IFS= read -r pr; do
        local pr_number pr_title pr_author pr_mergeable pr_checks pr_labels

        pr_number=$(echo "$pr" | jq -r '.number')
        pr_title=$(echo "$pr" | jq -r '.title')
        pr_author=$(echo "$pr" | jq -r '.user.login')
        pr_mergeable=$(echo "$pr" | jq -r '.mergeable')
        pr_labels=$(echo "$pr" | jq -r '.labels[].name' | tr '\n' ',' | sed 's/,$//')

        # Check if author is allowed
        if [[ -n "$allowed_actors" ]]; then
            local author_allowed=false
            while IFS= read -r allowed_actor; do
                if [[ "$pr_author" == "$allowed_actor" ]]; then
                    author_allowed=true
                    break
                fi
            done <<< "$allowed_actors"

            if [[ "$author_allowed" == false ]]; then
                continue
            fi
        fi

        # Check for skip labels
        if [[ -n "$skip_labels" ]] && [[ -n "$pr_labels" ]]; then
            local should_skip=false
            while IFS= read -r skip_label; do
                if [[ "$pr_labels" == *"$skip_label"* ]]; then
                    should_skip=true
                    break
                fi
            done <<< "$skip_labels"

            if [[ "$should_skip" == true ]]; then
                continue
            fi
        fi

        # Get status checks
        local checks_status="unknown"
        local checks_response
        checks_response=$(gh api "repos/$repo_name/pulls/$pr_number/checks" 2>/dev/null || echo '{"check_runs": []}')

        if [[ "$checks_response" != '{"check_runs": []}' ]]; then
            local total_checks success_checks
            total_checks=$(echo "$checks_response" | jq '.check_runs | length')
            success_checks=$(echo "$checks_response" | jq '[.check_runs[] | select(.conclusion == "success")] | length')

            if [[ "$total_checks" -eq 0 ]]; then
                checks_status="none"
            elif [[ "$success_checks" -eq "$total_checks" ]]; then
                checks_status="passing"
            else
                checks_status="failing"
            fi
        fi

        # Determine if PR is ready to merge
        local ready_to_merge="no"
        if [[ "$pr_mergeable" == "true" ]] && [[ "$checks_status" == "passing" ]]; then
            ready_to_merge="yes"
        fi

        # Output based on format
        if [[ "$OUTPUT_FORMAT" == "json" ]]; then
            jq -n \
                --arg repo "$repo_name" \
                --arg number "$pr_number" \
                --arg title "$pr_title" \
                --arg author "$pr_author" \
                --arg mergeable "$pr_mergeable" \
                --arg checks "$checks_status" \
                --arg ready "$ready_to_merge" \
                --arg labels "$pr_labels" \
                '{repository: $repo, pr_number: ($number | tonumber), title: $title, author: $author, mergeable: ($mergeable == "true"), checks_status: $checks, ready_to_merge: ($ready == "yes"), labels: ($labels | split(",") | map(select(. != "")))}'
        else
            printf "%-30s %-8s %-50s %-20s %-10s %-10s %-15s\n" \
                "$repo_name" \
                "#$pr_number" \
                "${pr_title:0:47}..." \
                "$pr_author" \
                "$pr_mergeable" \
                "$checks_status" \
                "$ready_to_merge"
        fi
    done
}

# Check GitLab PRs (Merge Requests)
check_gitlab_prs() {
    local repo_name="$1"
    local repo_url="$2"

    if [[ -z "$GITLAB_TOKEN" ]]; then
        log_error "GITLAB_TOKEN environment variable not set"
        return 1
    fi

    log_info "Checking GitLab repository: $repo_name"

    local project_id
    project_id=$(echo "$repo_name" | sed 's/\//%2F/g')

    # Get open merge requests
    local gitlab_url
    gitlab_url=$(yq '.auth.gitlab.url' "$CONFIG_FILE" 2>/dev/null || echo "https://gitlab.com")

    local mrs_json
    mrs_json=$(curl -s --header "PRIVATE-TOKEN: $GITLAB_TOKEN" \
        "$gitlab_url/api/v4/projects/$project_id/merge_requests?state=opened" || echo "[]")

    if [[ "$mrs_json" == "[]" ]] || [[ -z "$mrs_json" ]]; then
        echo "No open merge requests found for $repo_name"
        return 0
    fi

    # Process each MR
    echo "$mrs_json" | jq -c '.[]' | while IFS= read -r mr; do
        local mr_iid mr_title mr_author mr_mergeable mr_pipeline_status

        mr_iid=$(echo "$mr" | jq -r '.iid')
        mr_title=$(echo "$mr" | jq -r '.title')
        mr_author=$(echo "$mr" | jq -r '.author.username')
        mr_mergeable=$(echo "$mr" | jq -r '.merge_status')
        mr_pipeline_status=$(echo "$mr" | jq -r '.head_pipeline.status // "none"')

        # Determine if MR is ready to merge
        local ready_to_merge="no"
        if [[ "$mr_mergeable" == "can_be_merged" ]] && [[ "$mr_pipeline_status" == "success" ]]; then
            ready_to_merge="yes"
        fi

        # Output based on format
        if [[ "$OUTPUT_FORMAT" == "json" ]]; then
            jq -n \
                --arg repo "$repo_name" \
                --arg number "$mr_iid" \
                --arg title "$mr_title" \
                --arg author "$mr_author" \
                --arg mergeable "$mr_mergeable" \
                --arg pipeline "$mr_pipeline_status" \
                --arg ready "$ready_to_merge" \
                '{repository: $repo, pr_number: ($number | tonumber), title: $title, author: $author, mergeable: ($mergeable == "can_be_merged"), pipeline_status: $pipeline, ready_to_merge: ($ready == "yes")}'
        else
            printf "%-30s %-8s %-50s %-20s %-10s %-10s %-15s\n" \
                "$repo_name" \
                "!$mr_iid" \
                "${mr_title:0:47}..." \
                "$mr_author" \
                "$mr_mergeable" \
                "$mr_pipeline_status" \
                "$ready_to_merge"
        fi
    done
}

# Check Bitbucket PRs
check_bitbucket_prs() {
    local repo_name="$1"
    local repo_url="$2"

    if [[ -z "$BITBUCKET_USERNAME" ]] || [[ -z "$BITBUCKET_APP_PASSWORD" ]]; then
        log_error "BITBUCKET_USERNAME or BITBUCKET_APP_PASSWORD environment variable not set"
        return 1
    fi

    log_info "Checking Bitbucket repository: $repo_name"

    # Get open pull requests
    local prs_json
    prs_json=$(curl -s -u "$BITBUCKET_USERNAME:$BITBUCKET_APP_PASSWORD" \
        "https://api.bitbucket.org/2.0/repositories/$repo_name/pullrequests?state=OPEN" | jq '.values' || echo "[]")

    if [[ "$prs_json" == "[]" ]] || [[ -z "$prs_json" ]]; then
        echo "No open pull requests found for $repo_name"
        return 0
    fi

    # Process each PR
    echo "$prs_json" | jq -c '.[]' | while IFS= read -r pr; do
        local pr_id pr_title pr_author pr_state

        pr_id=$(echo "$pr" | jq -r '.id')
        pr_title=$(echo "$pr" | jq -r '.title')
        pr_author=$(echo "$pr" | jq -r '.author.username')
        pr_state=$(echo "$pr" | jq -r '.state')

        # For Bitbucket, we'll assume PRs are ready if they're open (limited API access)
        local ready_to_merge="unknown"

        # Output based on format
        if [[ "$OUTPUT_FORMAT" == "json" ]]; then
            jq -n \
                --arg repo "$repo_name" \
                --arg number "$pr_id" \
                --arg title "$pr_title" \
                --arg author "$pr_author" \
                --arg state "$pr_state" \
                --arg ready "$ready_to_merge" \
                '{repository: $repo, pr_number: ($number | tonumber), title: $title, author: $author, state: $state, ready_to_merge: $ready}'
        else
            printf "%-30s %-8s %-50s %-20s %-10s %-10s %-15s\n" \
                "$repo_name" \
                "#$pr_id" \
                "${pr_title:0:47}..." \
                "$pr_author" \
                "$pr_state" \
                "unknown" \
                "$ready_to_merge"
        fi
    done
}

# Main function
main() {
    log_info "Starting PR status check across multiple repositories"

    check_dependencies
    parse_config

    # Print table header for non-JSON output
    if [[ "$OUTPUT_FORMAT" != "json" ]]; then
        printf "%-30s %-8s %-50s %-20s %-10s %-10s %-15s\n" \
            "REPOSITORY" "PR" "TITLE" "AUTHOR" "MERGEABLE" "CHECKS" "READY_TO_MERGE"
        printf "%-30s %-8s %-50s %-20s %-10s %-10s %-15s\n" \
            "$(printf "%*s" 30 "" | tr " " "-")" \
            "$(printf "%*s" 8 "" | tr " " "-")" \
            "$(printf "%*s" 50 "" | tr " " "-")" \
            "$(printf "%*s" 20 "" | tr " " "-")" \
            "$(printf "%*s" 10 "" | tr " " "-")" \
            "$(printf "%*s" 10 "" | tr " " "-")" \
            "$(printf "%*s" 15 "" | tr " " "-")"
    fi

    # Process GitHub repositories
    if yq '.repositories.github' "$CONFIG_FILE" &> /dev/null; then
        yq '.repositories.github[] | .name + "," + .url' "$CONFIG_FILE" | while IFS=',' read -r repo_name repo_url; do
            check_github_prs "$repo_name" "$repo_url"
        done
    fi

    # Process GitLab repositories
    if yq '.repositories.gitlab' "$CONFIG_FILE" &> /dev/null; then
        yq '.repositories.gitlab[] | .name + "," + .url' "$CONFIG_FILE" | while IFS=',' read -r repo_name repo_url; do
            check_gitlab_prs "$repo_name" "$repo_url"
        done
    fi

    # Process Bitbucket repositories
    if yq '.repositories.bitbucket' "$CONFIG_FILE" &> /dev/null; then
        yq '.repositories.bitbucket[] | .name + "," + .url' "$CONFIG_FILE" | while IFS=',' read -r repo_name repo_url; do
            check_bitbucket_prs "$repo_name" "$repo_url"
        done
    fi

    log_success "PR status check completed"
}

# Show usage information
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Check PR status across multiple repositories defined in config.yaml

OPTIONS:
    -h, --help          Show this help message
    -c, --config FILE   Configuration file (default: config.yaml)
    -f, --format FORMAT Output format: table, json, csv (default: table)

ENVIRONMENT VARIABLES:
    GITHUB_TOKEN        GitHub personal access token
    GITLAB_TOKEN        GitLab personal access token
    BITBUCKET_USERNAME  Bitbucket username
    BITBUCKET_APP_PASSWORD Bitbucket app password

EXAMPLES:
    $0                                 # Check PRs with default settings
    $0 -f json                         # Output in JSON format
    $0 -c custom-config.yaml           # Use custom configuration file
    OUTPUT_FORMAT=json $0              # Set output format via environment

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -c|--config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        -f|--format)
            OUTPUT_FORMAT="$2"
            shift 2
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Run main function
main