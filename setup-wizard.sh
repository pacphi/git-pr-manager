#!/bin/bash

# setup-wizard.sh - Interactive configuration wizard for repository discovery
# This script helps users automatically discover and configure repositories across GitHub, GitLab, and Bitbucket

set -euo pipefail

# Configuration
CONFIG_FILE="${CONFIG_FILE:-config.yaml}"
BACKUP_DIR="backups"
TEMP_DIR="/tmp/multi-gitter-wizard"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Unicode symbols
CHECK_MARK="✓"
CROSS_MARK="✗"
INFO_MARK="ℹ"
WARNING_MARK="⚠"
ROCKET="🚀"
WRENCH="🔧"

# Log functions with consistent formatting
log_info() {
    echo -e "${BLUE}[${INFO_MARK} INFO]${NC} $1" >&2
}

log_success() {
    echo -e "${GREEN}[${CHECK_MARK} SUCCESS]${NC} $1" >&2
}

log_warning() {
    echo -e "${YELLOW}[${WARNING_MARK} WARNING]${NC} $1" >&2
}

log_error() {
    echo -e "${RED}[${CROSS_MARK} ERROR]${NC} $1" >&2
}

log_header() {
    echo -e "\n${BOLD}${CYAN}==== $1 ====${NC}" >&2
}

log_step() {
    echo -e "\n${PURPLE}${WRENCH} $1${NC}" >&2
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

# Check dependencies
check_dependencies() {
    local missing_deps=()

    if ! command -v yq &> /dev/null; then
        missing_deps+=("yq")
    fi

    if ! command -v jq &> /dev/null; then
        missing_deps+=("jq")
    fi

    if ! command -v curl &> /dev/null; then
        missing_deps+=("curl")
    fi

    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        log_info "Please run 'make install' to install required dependencies"
        exit 1
    fi
}

# Create temporary directory for wizard files
setup_temp_dir() {
    mkdir -p "$TEMP_DIR"

    # Cleanup temp directory on exit
    trap 'rm -rf "$TEMP_DIR"' EXIT
}

# Backup existing config if it exists
backup_config() {
    if [[ -f "$CONFIG_FILE" ]]; then
        local timestamp
        timestamp=$(date +%Y%m%d_%H%M%S)
        local backup_file="${CONFIG_FILE}.backup.${timestamp}"

        mkdir -p "$BACKUP_DIR"
        cp "$CONFIG_FILE" "$backup_file"
        log_success "Backed up existing config to: $backup_file"
        echo "$backup_file" > "$TEMP_DIR/backup_file"
    fi
}

# Check authentication tokens and validate access
check_authentication() {
    local providers=()
    local auth_status=()

    log_step "Checking authentication for Git providers..."

    # Check GitHub authentication
    if [[ -n "${GITHUB_TOKEN:-}" ]]; then
        if validate_github_auth; then
            providers+=("github")
            auth_status+=("${GREEN}GitHub: ${CHECK_MARK} Authenticated${NC}")
        else
            auth_status+=("${RED}GitHub: ${CROSS_MARK} Invalid token${NC}")
        fi
    else
        auth_status+=("${YELLOW}GitHub: ${WARNING_MARK} No token found${NC}")
    fi

    # Check GitLab authentication
    if [[ -n "${GITLAB_TOKEN:-}" ]]; then
        if validate_gitlab_auth; then
            providers+=("gitlab")
            auth_status+=("${GREEN}GitLab: ${CHECK_MARK} Authenticated${NC}")
        else
            auth_status+=("${RED}GitLab: ${CROSS_MARK} Invalid token${NC}")
        fi
    else
        auth_status+=("${YELLOW}GitLab: ${WARNING_MARK} No token found${NC}")
    fi

    # Check Bitbucket authentication
    if [[ -n "${BITBUCKET_USERNAME:-}" ]] && [[ -n "${BITBUCKET_APP_PASSWORD:-}" ]]; then
        if validate_bitbucket_auth; then
            providers+=("bitbucket")
            auth_status+=("${GREEN}Bitbucket: ${CHECK_MARK} Authenticated${NC}")
        else
            auth_status+=("${RED}Bitbucket: ${CROSS_MARK} Invalid credentials${NC}")
        fi
    else
        auth_status+=("${YELLOW}Bitbucket: ${WARNING_MARK} No credentials found${NC}")
    fi

    # Display authentication status
    echo ""
    for status in "${auth_status[@]}"; do
        echo -e "  $status"
    done
    echo ""

    # Save available providers to temp file
    printf '%s\n' "${providers[@]}" > "$TEMP_DIR/available_providers" 2>/dev/null || touch "$TEMP_DIR/available_providers"

    if [[ ${#providers[@]} -eq 0 ]]; then
        log_error "No valid authentication found for any Git provider"
        log_info "Please set the following environment variables:"
        echo -e "  ${CYAN}GitHub:${NC} export GITHUB_TOKEN=\"your_token_here\""
        echo -e "  ${CYAN}GitLab:${NC} export GITLAB_TOKEN=\"your_token_here\""
        echo -e "  ${CYAN}Bitbucket:${NC} export BITBUCKET_USERNAME=\"username\" && export BITBUCKET_APP_PASSWORD=\"password\""
        echo ""
        exit 1
    fi

    log_success "Found ${#providers[@]} authenticated provider(s): ${providers[*]}"
}

# Validate GitHub authentication
validate_github_auth() {
    local response
    response=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
        "https://api.github.com/user" 2>/dev/null || echo "")

    [[ -n "$response" ]] && echo "$response" | jq -e '.login' &>/dev/null
}

# Validate GitLab authentication
validate_gitlab_auth() {
    local gitlab_url response
    gitlab_url="${GITLAB_URL:-https://gitlab.com}"

    response=$(curl -s -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
        "$gitlab_url/api/v4/user" 2>/dev/null || echo "")

    [[ -n "$response" ]] && echo "$response" | jq -e '.username' &>/dev/null
}

# Validate Bitbucket authentication
validate_bitbucket_auth() {
    local response
    response=$(curl -s -u "$BITBUCKET_USERNAME:$BITBUCKET_APP_PASSWORD" \
        "https://api.bitbucket.org/2.0/user" 2>/dev/null || echo "")

    [[ -n "$response" ]] && echo "$response" | jq -e '.username' &>/dev/null
}

# Display welcome message
show_welcome() {
    clear
    echo -e "${BOLD}${CYAN}===============================================================================${NC}"
    echo -e "${BOLD}${CYAN}                    Multi-Gitter Configuration Wizard${NC}"
    echo ""
    echo -e "    ${ROCKET} Automatically discover repositories from your Git providers"
    echo -e "    ${WRENCH} Generate configuration with smart filtering and selection"
    echo -e "    ${CHECK_MARK} Setup PR automation in minutes, not hours"
    echo -e "${BOLD}${CYAN}===============================================================================${NC}"

    echo -e "\n${YELLOW}This wizard will help you:${NC}"
    echo "  1. Discover repositories from GitHub, GitLab, and Bitbucket"
    echo "  2. Filter and select repositories for PR automation"
    echo "  3. Configure merge strategies and auto-merge settings"
    echo "  4. Generate or update your config.yaml file"
    echo ""
}

# Show usage information
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Interactive wizard to discover repositories and generate configuration.

OPTIONS:
    -h, --help              Show this help message
    -c, --config FILE       Configuration file (default: config.yaml)
    -p, --preview           Preview mode - show what would be configured without making changes
    -a, --additive          Add to existing configuration instead of replacing it
    --backup-dir DIR        Directory for configuration backups (default: backups)

ENVIRONMENT VARIABLES:
    GITHUB_TOKEN            GitHub personal access token
    GITLAB_TOKEN            GitLab personal access token
    GITLAB_URL              GitLab instance URL (default: https://gitlab.com)
    BITBUCKET_USERNAME      Bitbucket username
    BITBUCKET_APP_PASSWORD  Bitbucket app password

EXAMPLES:
    $0                      # Run interactive wizard
    $0 --preview            # Preview what would be configured
    $0 --additive           # Add repositories to existing configuration
    $0 -c custom.yaml       # Use custom configuration file

For more information, see: https://github.com/your-org/multi-gitter-pr-automation
EOF
}

# Parse command line arguments
parse_arguments() {
    PREVIEW_MODE=false
    ADDITIVE_MODE=false

    # Create temp directory first, before we try to write to it
    mkdir -p "$TEMP_DIR"

    # Set up cleanup trap early since we're creating the temp directory
    trap 'rm -rf "$TEMP_DIR"' EXIT

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
            -p|--preview)
                PREVIEW_MODE=true
                shift
                ;;
            -a|--additive)
                ADDITIVE_MODE=true
                shift
                ;;
            --backup-dir)
                BACKUP_DIR="$2"
                shift 2
                ;;
            *)
                log_error "Unknown option: $1"
                usage >&2
                exit 1
                ;;
        esac
    done

    # Save settings to temp file
    echo "PREVIEW_MODE=$PREVIEW_MODE" > "$TEMP_DIR/settings"
    echo "ADDITIVE_MODE=$ADDITIVE_MODE" >> "$TEMP_DIR/settings"
    echo "CONFIG_FILE=$CONFIG_FILE" >> "$TEMP_DIR/settings"
}

# Discover repositories from all available providers
discover_repositories() {
    log_header "Repository Discovery"

    local available_providers
    available_providers=()
    if [[ -f "$TEMP_DIR/available_providers" ]]; then
        while IFS= read -r line || [[ -n "$line" ]]; do
            [[ -n "$line" ]] && available_providers+=("$line")
        done < "$TEMP_DIR/available_providers"
    fi

    if [[ ${#available_providers[@]} -eq 0 ]]; then
        log_error "No authenticated providers available"
        exit 1
    fi

    # Discover repositories for each provider
    for provider in "${available_providers[@]}"; do
        case "$provider" in
            "github")
                discover_github_repositories
                ;;
            "gitlab")
                discover_gitlab_repositories
                ;;
            "bitbucket")
                discover_bitbucket_repositories
                ;;
        esac
    done

    # Process discovered repositories
    process_repository_selections
}

# Discover GitHub repositories
discover_github_repositories() {
    log_step "Discovering GitHub repositories..."

    local user_info repos_file total_repos
    repos_file="$TEMP_DIR/github_repositories.json"

    # Get user information
    user_info=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
        "https://api.github.com/user" 2>/dev/null || echo "{}")

    local username
    username=$(echo "$user_info" | jq -r '.login // "unknown"')
    log_info "GitHub user: $username"

    # Discover repositories from multiple sources
    echo "[]" > "$repos_file"

    # 1. User's own repositories
    log_info "Fetching personal repositories..."
    fetch_github_user_repos "$username" "$repos_file"

    # 2. Organization repositories
    log_info "Fetching organization repositories..."
    fetch_github_org_repos "$repos_file"

    # Get final count
    total_repos=$(jq '. | length' "$repos_file")
    log_success "Discovered $total_repos GitHub repositories"

    # Apply filtering if repositories found
    if [[ "$total_repos" -gt 0 ]]; then
        filter_github_repositories "$repos_file"
    fi
}

# Fetch user's personal repositories
fetch_github_user_repos() {
    local username="$1"
    local repos_file="$2"
    local page=1
    local per_page=100

    while true; do
        local response
        response=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
            "https://api.github.com/user/repos?per_page=$per_page&page=$page&sort=updated" \
            2>/dev/null || echo "[]")

        # Check if response is an array and not empty
        if [[ "$response" == "[]" ]] || ! echo "$response" | jq -e 'type == "array" and length > 0' &>/dev/null; then
            # Check if it's an error response
            if echo "$response" | jq -e '.message' &>/dev/null; then
                local error_msg
                error_msg=$(echo "$response" | jq -r '.message')
                log_warning "GitHub API error: $error_msg"
            fi
            break
        fi

        # Merge with existing repositories
        jq --slurpfile existing "$repos_file" '$existing[0] + .' <<< "$response" > "${repos_file}.tmp"
        mv "${repos_file}.tmp" "$repos_file"

        ((page++))

        # GitHub API rate limiting protection
        sleep 0.1
    done
}

# Fetch organization repositories
fetch_github_org_repos() {
    local repos_file="$1"

    # Get user's organizations
    local orgs_response orgs
    orgs_response=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
        "https://api.github.com/user/orgs" 2>/dev/null || echo "[]")

    local orgs=()
    while IFS= read -r org || [[ -n "$org" ]]; do
        [[ -n "$org" && "$org" != "null" ]] && orgs+=("$org")
    done < <(echo "$orgs_response" | jq -r '.[] | .login' 2>/dev/null || true)

    for org in "${orgs[@]}"; do
        [[ -z "$org" ]] && continue

        log_info "Fetching repositories for organization: $org"
        local page=1
        local per_page=100

        while true; do
            local response
            response=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
                "https://api.github.com/orgs/$org/repos?per_page=$per_page&page=$page&sort=updated" \
                2>/dev/null || echo "[]")

            # Check if response is an array and not empty
            if [[ "$response" == "[]" ]] || ! echo "$response" | jq -e 'type == "array" and length > 0' &>/dev/null; then
                # Check if it's an error response
                if echo "$response" | jq -e '.message' &>/dev/null; then
                    local error_msg
                    error_msg=$(echo "$response" | jq -r '.message')
                    log_warning "GitHub API error for org $org: $error_msg"
                fi
                break
            fi

            # Merge with existing repositories, removing duplicates by URL
            jq --slurpfile existing "$repos_file" \
                '$existing[0] + . | unique_by(.clone_url)' <<< "$response" > "${repos_file}.tmp"
            mv "${repos_file}.tmp" "$repos_file"

            ((page++))
            sleep 0.1
        done
    done
}

# Filter GitHub repositories based on user preferences
filter_github_repositories() {
    local repos_file="$1"
    local filtered_file="$TEMP_DIR/github_repositories_filtered.json"

    log_step "Filtering GitHub repositories..."

    # Interactive filtering menu
    echo ""
    echo -e "${BOLD}Repository Filtering Options:${NC}"
    echo "1. Show all repositories"
    echo "2. Filter by visibility (public/private)"
    echo "3. Filter by owner (personal/organization)"
    echo "4. Filter by activity (last updated)"
    echo "5. Filter by name pattern"
    echo "6. Custom filters"
    echo ""

    read -p "Choose filtering option (1-6): " filter_choice

    case "$filter_choice" in
        1)
            # Show all repositories
            cp "$repos_file" "$filtered_file"
            ;;
        2)
            filter_by_visibility "$repos_file" "$filtered_file"
            ;;
        3)
            filter_by_owner "$repos_file" "$filtered_file"
            ;;
        4)
            filter_by_activity "$repos_file" "$filtered_file"
            ;;
        5)
            filter_by_name_pattern "$repos_file" "$filtered_file"
            ;;
        6)
            apply_custom_filters "$repos_file" "$filtered_file"
            ;;
        *)
            log_warning "Invalid choice, showing all repositories"
            cp "$repos_file" "$filtered_file"
            ;;
    esac

    local filtered_count
    filtered_count=$(jq '. | length' "$filtered_file")
    log_success "Filtered to $filtered_count repositories"

    # Save filtered results
    mv "$filtered_file" "$repos_file"
}

# Filter repositories by visibility (public/private)
filter_by_visibility() {
    local input_file="$1"
    local output_file="$2"

    echo ""
    echo "Filter by visibility:"
    echo "1. Public repositories only"
    echo "2. Private repositories only"
    echo "3. Both public and private"
    echo ""

    read -p "Choose visibility (1-3): " visibility_choice

    case "$visibility_choice" in
        1)
            jq '[.[] | select(.private == false)]' "$input_file" > "$output_file"
            ;;
        2)
            jq '[.[] | select(.private == true)]' "$input_file" > "$output_file"
            ;;
        3)
            cp "$input_file" "$output_file"
            ;;
        *)
            log_warning "Invalid choice, including all repositories"
            cp "$input_file" "$output_file"
            ;;
    esac
}

# Filter repositories by owner type
filter_by_owner() {
    local input_file="$1"
    local output_file="$2"

    echo ""
    echo "Filter by owner type:"
    echo "1. Personal repositories only"
    echo "2. Organization repositories only"
    echo "3. Both personal and organization"
    echo ""

    read -p "Choose owner type (1-3): " owner_choice

    case "$owner_choice" in
        1)
            jq '[.[] | select(.owner.type == "User")]' "$input_file" > "$output_file"
            ;;
        2)
            jq '[.[] | select(.owner.type == "Organization")]' "$input_file" > "$output_file"
            ;;
        3)
            cp "$input_file" "$output_file"
            ;;
        *)
            log_warning "Invalid choice, including all repositories"
            cp "$input_file" "$output_file"
            ;;
    esac
}

# Filter repositories by activity (last updated)
filter_by_activity() {
    local input_file="$1"
    local output_file="$2"

    echo ""
    echo "Filter by last activity:"
    echo "1. Updated within last 30 days"
    echo "2. Updated within last 90 days"
    echo "3. Updated within last 365 days"
    echo "4. No activity filter"
    echo ""

    read -p "Choose activity filter (1-4): " activity_choice

    local cutoff_date
    case "$activity_choice" in
        1)
            cutoff_date=$(date -d '30 days ago' '+%Y-%m-%d' 2>/dev/null || date -v-30d '+%Y-%m-%d')
            ;;
        2)
            cutoff_date=$(date -d '90 days ago' '+%Y-%m-%d' 2>/dev/null || date -v-90d '+%Y-%m-%d')
            ;;
        3)
            cutoff_date=$(date -d '365 days ago' '+%Y-%m-%d' 2>/dev/null || date -v-365d '+%Y-%m-%d')
            ;;
        4)
            cp "$input_file" "$output_file"
            return
            ;;
        *)
            log_warning "Invalid choice, no activity filter applied"
            cp "$input_file" "$output_file"
            return
            ;;
    esac

    jq --arg cutoff "$cutoff_date" '[.[] | select(.updated_at >= $cutoff)]' "$input_file" > "$output_file"
}

# Filter repositories by name pattern
filter_by_name_pattern() {
    local input_file="$1"
    local output_file="$2"

    echo ""
    read -p "Enter name pattern to match (e.g., 'frontend', 'api-*', '*-service'): " name_pattern

    if [[ -z "$name_pattern" ]]; then
        log_warning "No pattern provided, including all repositories"
        cp "$input_file" "$output_file"
        return
    fi

    # Convert shell glob pattern to regex
    local regex_pattern
    regex_pattern=$(echo "$name_pattern" | sed 's/\*/.*\?/g')

    jq --arg pattern "$regex_pattern" '[.[] | select(.name | test($pattern; "i"))]' "$input_file" > "$output_file"
}

# Apply custom filters
apply_custom_filters() {
    local input_file="$1"
    local output_file="$2"

    cp "$input_file" "$output_file"

    echo ""
    echo -e "${BOLD}Custom Filters:${NC}"

    # Fork filter
    echo ""
    read -p "Include forked repositories? (y/n): " include_forks
    if [[ "$include_forks" =~ ^[Nn] ]]; then
        jq '[.[] | select(.fork == false)]' "$output_file" > "${output_file}.tmp"
        mv "${output_file}.tmp" "$output_file"
    fi

    # Archived filter
    echo ""
    read -p "Include archived repositories? (y/n): " include_archived
    if [[ "$include_archived" =~ ^[Nn] ]]; then
        jq '[.[] | select(.archived == false)]' "$output_file" > "${output_file}.tmp"
        mv "${output_file}.tmp" "$output_file"
    fi

    # Language filter
    echo ""
    read -p "Filter by primary language (leave empty for all): " language_filter
    if [[ -n "$language_filter" ]]; then
        jq --arg lang "$language_filter" '[.[] | select(.language == $lang)]' "$output_file" > "${output_file}.tmp"
        mv "${output_file}.tmp" "$output_file"
    fi
}

# Discover GitLab repositories
discover_gitlab_repositories() {
    log_step "Discovering GitLab repositories..."

    local user_info repos_file total_repos gitlab_url
    repos_file="$TEMP_DIR/gitlab_repositories.json"
    gitlab_url="${GITLAB_URL:-https://gitlab.com}"

    # Get user information
    user_info=$(curl -s -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
        "$gitlab_url/api/v4/user" 2>/dev/null || echo "{}")

    local username
    username=$(echo "$user_info" | jq -r '.username // "unknown"')
    log_info "GitLab user: $username"

    # Discover repositories from multiple sources
    echo "[]" > "$repos_file"

    # 1. User's personal projects
    log_info "Fetching personal projects..."
    fetch_gitlab_user_projects "$repos_file" "$gitlab_url"

    # 2. Group projects
    log_info "Fetching group projects..."
    fetch_gitlab_group_projects "$repos_file" "$gitlab_url"

    # Get final count
    total_repos=$(jq '. | length' "$repos_file")
    log_success "Discovered $total_repos GitLab projects"

    # Apply filtering if repositories found
    if [[ "$total_repos" -gt 0 ]]; then
        filter_gitlab_repositories "$repos_file"
    fi
}

# Fetch user's personal GitLab projects
fetch_gitlab_user_projects() {
    local repos_file="$1"
    local gitlab_url="$2"
    local page=1
    local per_page=100

    while true; do
        local response
        response=$(curl -s -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
            "$gitlab_url/api/v4/projects?owned=true&per_page=$per_page&page=$page&order_by=updated_at&sort=desc" \
            2>/dev/null || echo "[]")

        if [[ "$response" == "[]" ]] || ! echo "$response" | jq -e '. | length > 0' &>/dev/null; then
            break
        fi

        # Merge with existing repositories
        jq --slurpfile existing "$repos_file" '$existing[0] + .' <<< "$response" > "${repos_file}.tmp"
        mv "${repos_file}.tmp" "$repos_file"

        ((page++))

        # GitLab API rate limiting protection
        sleep 0.1
    done
}

# Fetch GitLab group projects
fetch_gitlab_group_projects() {
    local repos_file="$1"
    local gitlab_url="$2"

    # Get user's groups
    local groups_response groups
    groups_response=$(curl -s -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
        "$gitlab_url/api/v4/groups?min_access_level=20" 2>/dev/null || echo "[]")

    local groups=()
    while IFS= read -r group_id || [[ -n "$group_id" ]]; do
        [[ -n "$group_id" && "$group_id" != "null" ]] && groups+=("$group_id")
    done < <(echo "$groups_response" | jq -r '.[] | .id' 2>/dev/null || true)

    for group_id in "${groups[@]}"; do
        [[ -z "$group_id" ]] && continue

        # Get group name for logging
        local group_name
        group_name=$(echo "$groups_response" | jq -r ".[] | select(.id == $group_id) | .full_path")
        log_info "Fetching projects for group: $group_name"

        local page=1
        local per_page=100

        while true; do
            local response
            response=$(curl -s -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
                "$gitlab_url/api/v4/groups/$group_id/projects?per_page=$per_page&page=$page&order_by=updated_at&sort=desc" \
                2>/dev/null || echo "[]")

            if [[ "$response" == "[]" ]] || ! echo "$response" | jq -e '. | length > 0' &>/dev/null; then
                break
            fi

            # Merge with existing repositories, removing duplicates by URL
            jq --slurpfile existing "$repos_file" \
                '$existing[0] + . | unique_by(.http_url_to_repo)' <<< "$response" > "${repos_file}.tmp"
            mv "${repos_file}.tmp" "$repos_file"

            ((page++))
            sleep 0.1
        done
    done
}

# Filter GitLab repositories based on user preferences
filter_gitlab_repositories() {
    local repos_file="$1"
    local filtered_file="$TEMP_DIR/gitlab_repositories_filtered.json"

    log_step "Filtering GitLab projects..."

    # Interactive filtering menu
    echo ""
    echo -e "${BOLD}Project Filtering Options:${NC}"
    echo "1. Show all projects"
    echo "2. Filter by visibility (public/private/internal)"
    echo "3. Filter by namespace (personal/group)"
    echo "4. Filter by activity (last updated)"
    echo "5. Filter by name pattern"
    echo "6. Custom filters"
    echo ""

    read -r -p "Choose filtering option (1-6): " filter_choice

    case "$filter_choice" in
        1)
            # Show all projects
            cp "$repos_file" "$filtered_file"
            ;;
        2)
            filter_gitlab_by_visibility "$repos_file" "$filtered_file"
            ;;
        3)
            filter_gitlab_by_namespace "$repos_file" "$filtered_file"
            ;;
        4)
            filter_gitlab_by_activity "$repos_file" "$filtered_file"
            ;;
        5)
            filter_gitlab_by_name_pattern "$repos_file" "$filtered_file"
            ;;
        6)
            apply_gitlab_custom_filters "$repos_file" "$filtered_file"
            ;;
        *)
            log_warning "Invalid choice, showing all projects"
            cp "$repos_file" "$filtered_file"
            ;;
    esac

    local filtered_count
    filtered_count=$(jq '. | length' "$filtered_file")
    log_success "Filtered to $filtered_count projects"

    # Save filtered results
    mv "$filtered_file" "$repos_file"
}

# Filter GitLab projects by visibility
filter_gitlab_by_visibility() {
    local input_file="$1"
    local output_file="$2"

    echo ""
    echo "Filter by visibility:"
    echo "1. Public projects only"
    echo "2. Private projects only"
    echo "3. Internal projects only"
    echo "4. All visibility levels"
    echo ""

    read -r -p "Choose visibility (1-4): " visibility_choice

    case "$visibility_choice" in
        1)
            jq '[.[] | select(.visibility == "public")]' "$input_file" > "$output_file"
            ;;
        2)
            jq '[.[] | select(.visibility == "private")]' "$input_file" > "$output_file"
            ;;
        3)
            jq '[.[] | select(.visibility == "internal")]' "$input_file" > "$output_file"
            ;;
        4)
            cp "$input_file" "$output_file"
            ;;
        *)
            log_warning "Invalid choice, including all projects"
            cp "$input_file" "$output_file"
            ;;
    esac
}

# Filter GitLab projects by namespace type
filter_gitlab_by_namespace() {
    local input_file="$1"
    local output_file="$2"

    echo ""
    echo "Filter by namespace type:"
    echo "1. Personal projects only"
    echo "2. Group projects only"
    echo "3. Both personal and group"
    echo ""

    read -r -p "Choose namespace type (1-3): " namespace_choice

    case "$namespace_choice" in
        1)
            jq '[.[] | select(.namespace.kind == "user")]' "$input_file" > "$output_file"
            ;;
        2)
            jq '[.[] | select(.namespace.kind == "group")]' "$input_file" > "$output_file"
            ;;
        3)
            cp "$input_file" "$output_file"
            ;;
        *)
            log_warning "Invalid choice, including all projects"
            cp "$input_file" "$output_file"
            ;;
    esac
}

# Filter GitLab projects by activity
filter_gitlab_by_activity() {
    local input_file="$1"
    local output_file="$2"

    echo ""
    echo "Filter by last activity:"
    echo "1. Updated within last 30 days"
    echo "2. Updated within last 90 days"
    echo "3. Updated within last 365 days"
    echo "4. No activity filter"
    echo ""

    read -r -p "Choose activity filter (1-4): " activity_choice

    local cutoff_date
    case "$activity_choice" in
        1)
            cutoff_date=$(date -d '30 days ago' '+%Y-%m-%d' 2>/dev/null || date -v-30d '+%Y-%m-%d')
            ;;
        2)
            cutoff_date=$(date -d '90 days ago' '+%Y-%m-%d' 2>/dev/null || date -v-90d '+%Y-%m-%d')
            ;;
        3)
            cutoff_date=$(date -d '365 days ago' '+%Y-%m-%d' 2>/dev/null || date -v-365d '+%Y-%m-%d')
            ;;
        4)
            cp "$input_file" "$output_file"
            return
            ;;
        *)
            log_warning "Invalid choice, no activity filter applied"
            cp "$input_file" "$output_file"
            return
            ;;
    esac

    jq --arg cutoff "$cutoff_date" '[.[] | select(.last_activity_at >= $cutoff)]' "$input_file" > "$output_file"
}

# Filter GitLab projects by name pattern
filter_gitlab_by_name_pattern() {
    local input_file="$1"
    local output_file="$2"

    echo ""
    read -r -p "Enter name pattern to match (e.g., 'frontend', 'api-*', '*-service'): " name_pattern

    if [[ -z "$name_pattern" ]]; then
        log_warning "No pattern provided, including all projects"
        cp "$input_file" "$output_file"
        return
    fi

    # Convert shell glob pattern to regex
    local regex_pattern
    regex_pattern="${name_pattern//\*/.*?}"

    jq --arg pattern "$regex_pattern" '[.[] | select(.name | test($pattern; "i"))]' "$input_file" > "$output_file"
}

# Apply custom filters to GitLab projects
apply_gitlab_custom_filters() {
    local input_file="$1"
    local output_file="$2"

    cp "$input_file" "$output_file"

    echo ""
    echo -e "${BOLD}Custom Filters:${NC}"

    # Fork filter
    echo ""
    read -r -p "Include forked projects? (y/n): " include_forks
    if [[ "$include_forks" =~ ^[Nn] ]]; then
        jq '[.[] | select(.forked_from_project == null)]' "$output_file" > "${output_file}.tmp"
        mv "${output_file}.tmp" "$output_file"
    fi

    # Archived filter
    echo ""
    read -r -p "Include archived projects? (y/n): " include_archived
    if [[ "$include_archived" =~ ^[Nn] ]]; then
        jq '[.[] | select(.archived == false)]' "$output_file" > "${output_file}.tmp"
        mv "${output_file}.tmp" "$output_file"
    fi

    # Empty project filter
    echo ""
    read -r -p "Include empty projects (no commits)? (y/n): " include_empty
    if [[ "$include_empty" =~ ^[Nn] ]]; then
        jq '[.[] | select(.empty_repo == false)]' "$output_file" > "${output_file}.tmp"
        mv "${output_file}.tmp" "$output_file"
    fi
}

# Discover Bitbucket repositories
discover_bitbucket_repositories() {
    log_step "Discovering Bitbucket repositories..."

    local user_info repos_file total_repos
    repos_file="$TEMP_DIR/bitbucket_repositories.json"

    # Get user information
    user_info=$(curl -s -u "$BITBUCKET_USERNAME:$BITBUCKET_APP_PASSWORD" \
        "https://api.bitbucket.org/2.0/user" 2>/dev/null || echo "{}")

    local username
    username=$(echo "$user_info" | jq -r '.username // "unknown"')
    log_info "Bitbucket user: $username"

    # Discover repositories from multiple sources
    echo "[]" > "$repos_file"

    # 1. User's repositories
    log_info "Fetching personal repositories..."
    fetch_bitbucket_user_repos "$repos_file"

    # 2. Workspace repositories (teams)
    log_info "Fetching workspace repositories..."
    fetch_bitbucket_workspace_repos "$repos_file"

    # Get final count
    total_repos=$(jq '. | length' "$repos_file")
    log_success "Discovered $total_repos Bitbucket repositories"

    # Apply filtering if repositories found
    if [[ "$total_repos" -gt 0 ]]; then
        filter_bitbucket_repositories "$repos_file"
    fi
}

# Fetch user's personal Bitbucket repositories
fetch_bitbucket_user_repos() {
    local repos_file="$1"
    local page=1
    local pagelen=100

    while true; do
        local response
        response=$(curl -s -u "$BITBUCKET_USERNAME:$BITBUCKET_APP_PASSWORD" \
            "https://api.bitbucket.org/2.0/repositories/$BITBUCKET_USERNAME?pagelen=$pagelen&page=$page&sort=-updated_on" \
            2>/dev/null || echo '{"values": []}')

        local values
        values=$(echo "$response" | jq '.values // []')

        if [[ "$values" == "[]" ]] || ! echo "$values" | jq -e '. | length > 0' &>/dev/null; then
            break
        fi

        # Merge with existing repositories
        jq --slurpfile existing "$repos_file" '$existing[0] + .' <<< "$values" > "${repos_file}.tmp"
        mv "${repos_file}.tmp" "$repos_file"

        ((page++))

        # Bitbucket API rate limiting protection
        sleep 0.1
    done
}

# Fetch Bitbucket workspace repositories
fetch_bitbucket_workspace_repos() {
    local repos_file="$1"

    # Get user's workspaces
    local workspaces_response workspaces
    workspaces_response=$(curl -s -u "$BITBUCKET_USERNAME:$BITBUCKET_APP_PASSWORD" \
        "https://api.bitbucket.org/2.0/workspaces?role=member" 2>/dev/null || echo '{"values": []}')

    local workspaces=()
    while IFS= read -r workspace || [[ -n "$workspace" ]]; do
        [[ -n "$workspace" && "$workspace" != "null" ]] && workspaces+=("$workspace")
    done < <(echo "$workspaces_response" | jq -r '.values[]? | .slug' 2>/dev/null || true)

    for workspace in "${workspaces[@]}"; do
        [[ -z "$workspace" || "$workspace" == "$BITBUCKET_USERNAME" ]] && continue

        log_info "Fetching repositories for workspace: $workspace"
        local page=1
        local pagelen=100

        while true; do
            local response
            response=$(curl -s -u "$BITBUCKET_USERNAME:$BITBUCKET_APP_PASSWORD" \
                "https://api.bitbucket.org/2.0/repositories/$workspace?pagelen=$pagelen&page=$page&sort=-updated_on" \
                2>/dev/null || echo '{"values": []}')

            local values
            values=$(echo "$response" | jq '.values // []')

            if [[ "$values" == "[]" ]] || ! echo "$values" | jq -e '. | length > 0' &>/dev/null; then
                break
            fi

            # Merge with existing repositories, removing duplicates by URL
            jq --slurpfile existing "$repos_file" \
                '$existing[0] + . | unique_by(.links.clone[] | select(.name == "https") | .href)' <<< "$values" > "${repos_file}.tmp"
            mv "${repos_file}.tmp" "$repos_file"

            ((page++))
            sleep 0.1
        done
    done
}

# Filter Bitbucket repositories based on user preferences
filter_bitbucket_repositories() {
    local repos_file="$1"
    local filtered_file="$TEMP_DIR/bitbucket_repositories_filtered.json"

    log_step "Filtering Bitbucket repositories..."

    # Interactive filtering menu
    echo ""
    echo -e "${BOLD}Repository Filtering Options:${NC}"
    echo "1. Show all repositories"
    echo "2. Filter by access level (public/private)"
    echo "3. Filter by owner (personal/workspace)"
    echo "4. Filter by activity (last updated)"
    echo "5. Filter by name pattern"
    echo "6. Custom filters"
    echo ""

    read -r -p "Choose filtering option (1-6): " filter_choice

    case "$filter_choice" in
        1)
            # Show all repositories
            cp "$repos_file" "$filtered_file"
            ;;
        2)
            filter_bitbucket_by_access "$repos_file" "$filtered_file"
            ;;
        3)
            filter_bitbucket_by_owner "$repos_file" "$filtered_file"
            ;;
        4)
            filter_bitbucket_by_activity "$repos_file" "$filtered_file"
            ;;
        5)
            filter_bitbucket_by_name_pattern "$repos_file" "$filtered_file"
            ;;
        6)
            apply_bitbucket_custom_filters "$repos_file" "$filtered_file"
            ;;
        *)
            log_warning "Invalid choice, showing all repositories"
            cp "$repos_file" "$filtered_file"
            ;;
    esac

    local filtered_count
    filtered_count=$(jq '. | length' "$filtered_file")
    log_success "Filtered to $filtered_count repositories"

    # Save filtered results
    mv "$filtered_file" "$repos_file"
}

# Filter Bitbucket repositories by access level
filter_bitbucket_by_access() {
    local input_file="$1"
    local output_file="$2"

    echo ""
    echo "Filter by access level:"
    echo "1. Public repositories only"
    echo "2. Private repositories only"
    echo "3. Both public and private"
    echo ""

    read -r -p "Choose access level (1-3): " access_choice

    case "$access_choice" in
        1)
            jq '[.[] | select(.is_private == false)]' "$input_file" > "$output_file"
            ;;
        2)
            jq '[.[] | select(.is_private == true)]' "$input_file" > "$output_file"
            ;;
        3)
            cp "$input_file" "$output_file"
            ;;
        *)
            log_warning "Invalid choice, including all repositories"
            cp "$input_file" "$output_file"
            ;;
    esac
}

# Filter Bitbucket repositories by owner type
filter_bitbucket_by_owner() {
    local input_file="$1"
    local output_file="$2"

    echo ""
    echo "Filter by owner type:"
    echo "1. Personal repositories only"
    echo "2. Workspace repositories only"
    echo "3. Both personal and workspace"
    echo ""

    read -r -p "Choose owner type (1-3): " owner_choice

    case "$owner_choice" in
        1)
            jq --arg user "$BITBUCKET_USERNAME" '[.[] | select(.owner.username == $user)]' "$input_file" > "$output_file"
            ;;
        2)
            jq --arg user "$BITBUCKET_USERNAME" '[.[] | select(.owner.username != $user)]' "$input_file" > "$output_file"
            ;;
        3)
            cp "$input_file" "$output_file"
            ;;
        *)
            log_warning "Invalid choice, including all repositories"
            cp "$input_file" "$output_file"
            ;;
    esac
}

# Filter Bitbucket repositories by activity
filter_bitbucket_by_activity() {
    local input_file="$1"
    local output_file="$2"

    echo ""
    echo "Filter by last activity:"
    echo "1. Updated within last 30 days"
    echo "2. Updated within last 90 days"
    echo "3. Updated within last 365 days"
    echo "4. No activity filter"
    echo ""

    read -r -p "Choose activity filter (1-4): " activity_choice

    local cutoff_date
    case "$activity_choice" in
        1)
            cutoff_date=$(date -d '30 days ago' '+%Y-%m-%d' 2>/dev/null || date -v-30d '+%Y-%m-%d')
            ;;
        2)
            cutoff_date=$(date -d '90 days ago' '+%Y-%m-%d' 2>/dev/null || date -v-90d '+%Y-%m-%d')
            ;;
        3)
            cutoff_date=$(date -d '365 days ago' '+%Y-%m-%d' 2>/dev/null || date -v-365d '+%Y-%m-%d')
            ;;
        4)
            cp "$input_file" "$output_file"
            return
            ;;
        *)
            log_warning "Invalid choice, no activity filter applied"
            cp "$input_file" "$output_file"
            return
            ;;
    esac

    jq --arg cutoff "$cutoff_date" '[.[] | select(.updated_on >= $cutoff)]' "$input_file" > "$output_file"
}

# Filter Bitbucket repositories by name pattern
filter_bitbucket_by_name_pattern() {
    local input_file="$1"
    local output_file="$2"

    echo ""
    read -r -p "Enter name pattern to match (e.g., 'frontend', 'api-*', '*-service'): " name_pattern

    if [[ -z "$name_pattern" ]]; then
        log_warning "No pattern provided, including all repositories"
        cp "$input_file" "$output_file"
        return
    fi

    # Convert shell glob pattern to regex
    local regex_pattern
    regex_pattern="${name_pattern//\*/.*?}"

    jq --arg pattern "$regex_pattern" '[.[] | select(.name | test($pattern; "i"))]' "$input_file" > "$output_file"
}

# Apply custom filters to Bitbucket repositories
apply_bitbucket_custom_filters() {
    local input_file="$1"
    local output_file="$2"

    cp "$input_file" "$output_file"

    echo ""
    echo -e "${BOLD}Custom Filters:${NC}"

    # Fork filter
    echo ""
    read -r -p "Include forked repositories? (y/n): " include_forks
    if [[ "$include_forks" =~ ^[Nn] ]]; then
        jq '[.[] | select(.parent == null)]' "$output_file" > "${output_file}.tmp"
        mv "${output_file}.tmp" "$output_file"
    fi

    # Language filter
    echo ""
    read -r -p "Filter by primary language (leave empty for all): " language_filter
    if [[ -n "$language_filter" ]]; then
        jq --arg lang "$language_filter" '[.[] | select(.language == $lang)]' "$output_file" > "${output_file}.tmp"
        mv "${output_file}.tmp" "$output_file"
    fi
}

# Process repository selections (interactive menu)
process_repository_selections() {
    log_step "Repository Selection and Configuration"

    # Display discovery summary
    echo ""
    echo -e "${BOLD}${CYAN}Repository Discovery Summary:${NC}"

    local total_repos=0

    # GitHub summary
    if [[ -f "$TEMP_DIR/github_repositories.json" ]]; then
        local github_count
        github_count=$(jq '. | length' "$TEMP_DIR/github_repositories.json")
        echo -e "  ${GREEN}GitHub:${NC} $github_count repositories found"

        if [[ "$github_count" -gt 0 ]]; then
            echo "    Sample repositories:"
            jq -r '.[:3] | .[] | "      - " + .full_name + " (" + (.language // "Unknown") + ")"' "$TEMP_DIR/github_repositories.json" 2>/dev/null | head -3
            if [[ "$github_count" -gt 3 ]]; then
                echo "      ... and $((github_count - 3)) more"
            fi
        fi
        total_repos=$((total_repos + github_count))
        echo ""
    fi

    # GitLab summary
    if [[ -f "$TEMP_DIR/gitlab_repositories.json" ]]; then
        local gitlab_count
        gitlab_count=$(jq '. | length' "$TEMP_DIR/gitlab_repositories.json")
        echo -e "  ${GREEN}GitLab:${NC} $gitlab_count projects found"

        if [[ "$gitlab_count" -gt 0 ]]; then
            echo "    Sample projects:"
            jq -r '.[:3] | .[] | "      - " + .path_with_namespace + " (" + (.default_branch // "main") + ")"' "$TEMP_DIR/gitlab_repositories.json" 2>/dev/null | head -3
            if [[ "$gitlab_count" -gt 3 ]]; then
                echo "      ... and $((gitlab_count - 3)) more"
            fi
        fi
        total_repos=$((total_repos + gitlab_count))
        echo ""
    fi

    # Bitbucket summary
    if [[ -f "$TEMP_DIR/bitbucket_repositories.json" ]]; then
        local bitbucket_count
        bitbucket_count=$(jq '. | length' "$TEMP_DIR/bitbucket_repositories.json")
        echo -e "  ${GREEN}Bitbucket:${NC} $bitbucket_count repositories found"

        if [[ "$bitbucket_count" -gt 0 ]]; then
            echo "    Sample repositories:"
            jq -r '.[:3] | .[] | "      - " + .full_name + " (" + (.language // "Unknown") + ")"' "$TEMP_DIR/bitbucket_repositories.json" 2>/dev/null | head -3
            if [[ "$bitbucket_count" -gt 3 ]]; then
                echo "      ... and $((bitbucket_count - 3)) more"
            fi
        fi
        total_repos=$((total_repos + bitbucket_count))
        echo ""
    fi

    log_success "Total discovered: $total_repos repositories"

    # Interactive selection menu
    if [[ "$total_repos" -gt 0 ]]; then
        show_repository_selection_menu
    else
        log_warning "No repositories found to configure"
        return 1
    fi
}

# Show interactive repository selection menu
show_repository_selection_menu() {
    echo ""
    echo -e "${BOLD}${YELLOW}Repository Configuration Options:${NC}"
    echo "1. Select all repositories for auto-merge"
    echo "2. Select repositories by provider"
    echo "3. Select repositories interactively"
    echo "4. Configure merge strategies"
    echo "5. Generate configuration and exit"
    echo "6. Preview configuration"
    echo "0. Exit without saving"
    echo ""

    while true; do
        read -r -p "Choose an option (0-6): " choice

        case "$choice" in
            1)
                select_all_repositories
                break
                ;;
            2)
                select_by_provider
                break
                ;;
            3)
                select_repositories_interactively
                break
                ;;
            4)
                configure_merge_strategies
                continue
                ;;
            5)
                generate_configuration
                break
                ;;
            6)
                preview_configuration
                continue
                ;;
            0)
                log_info "Exiting without saving configuration"
                exit 0
                ;;
            *)
                log_warning "Invalid choice. Please select 0-6."
                continue
                ;;
        esac
    done
}

# Select all repositories for configuration
select_all_repositories() {
    log_step "Selecting all repositories for auto-merge..."

    # Create a combined selection file
    echo '{"repositories": {"github": [], "gitlab": [], "bitbucket": []}}' > "$TEMP_DIR/selected_repos.json"

    # Add GitHub repositories
    if [[ -f "$TEMP_DIR/github_repositories.json" ]]; then
        local github_count
        github_count=$(jq '. | length' "$TEMP_DIR/github_repositories.json")
        if [[ "$github_count" -gt 0 ]]; then
            jq --slurpfile github "$TEMP_DIR/github_repositories.json" \
                '.repositories.github = $github[0] | map_values(if type == "array" then . else . end)' \
                "$TEMP_DIR/selected_repos.json" > "$TEMP_DIR/selected_repos.json.tmp"
            mv "$TEMP_DIR/selected_repos.json.tmp" "$TEMP_DIR/selected_repos.json"
        fi
    fi

    # Add GitLab repositories
    if [[ -f "$TEMP_DIR/gitlab_repositories.json" ]]; then
        local gitlab_count
        gitlab_count=$(jq '. | length' "$TEMP_DIR/gitlab_repositories.json")
        if [[ "$gitlab_count" -gt 0 ]]; then
            jq --slurpfile gitlab "$TEMP_DIR/gitlab_repositories.json" \
                '.repositories.gitlab = $gitlab[0] | map_values(if type == "array" then . else . end)' \
                "$TEMP_DIR/selected_repos.json" > "$TEMP_DIR/selected_repos.json.tmp"
            mv "$TEMP_DIR/selected_repos.json.tmp" "$TEMP_DIR/selected_repos.json"
        fi
    fi

    # Add Bitbucket repositories
    if [[ -f "$TEMP_DIR/bitbucket_repositories.json" ]]; then
        local bitbucket_count
        bitbucket_count=$(jq '. | length' "$TEMP_DIR/bitbucket_repositories.json")
        if [[ "$bitbucket_count" -gt 0 ]]; then
            jq --slurpfile bitbucket "$TEMP_DIR/bitbucket_repositories.json" \
                '.repositories.bitbucket = $bitbucket[0] | map_values(if type == "array" then . else . end)' \
                "$TEMP_DIR/selected_repos.json" > "$TEMP_DIR/selected_repos.json.tmp"
            mv "$TEMP_DIR/selected_repos.json.tmp" "$TEMP_DIR/selected_repos.json"
        fi
    fi

    local total_selected
    total_selected=$(jq '.repositories.github + .repositories.gitlab + .repositories.bitbucket | length' "$TEMP_DIR/selected_repos.json")
    log_success "Selected $total_selected repositories for configuration"

    generate_configuration
}

# Select repositories by provider
select_by_provider() {
    log_step "Selecting repositories by provider..."

    echo ""
    echo "Available providers:"

    local providers=()
    if [[ -f "$TEMP_DIR/github_repositories.json" ]] && [[ $(jq '. | length' "$TEMP_DIR/github_repositories.json") -gt 0 ]]; then
        providers+=("github")
        echo "  1. GitHub ($(jq '. | length' "$TEMP_DIR/github_repositories.json") repositories)"
    fi

    if [[ -f "$TEMP_DIR/gitlab_repositories.json" ]] && [[ $(jq '. | length' "$TEMP_DIR/gitlab_repositories.json") -gt 0 ]]; then
        providers+=("gitlab")
        echo "  2. GitLab ($(jq '. | length' "$TEMP_DIR/gitlab_repositories.json") projects)"
    fi

    if [[ -f "$TEMP_DIR/bitbucket_repositories.json" ]] && [[ $(jq '. | length' "$TEMP_DIR/bitbucket_repositories.json") -gt 0 ]]; then
        providers+=("bitbucket")
        echo "  3. Bitbucket ($(jq '. | length' "$TEMP_DIR/bitbucket_repositories.json") repositories)"
    fi

    echo ""
    read -r -p "Select providers to include (comma-separated, e.g., 1,2): " provider_choices

    # Initialize selected repositories file
    echo '{"repositories": {"github": [], "gitlab": [], "bitbucket": []}}' > "$TEMP_DIR/selected_repos.json"

    # Process selections
    IFS=',' read -ra CHOICES <<< "$provider_choices"
    for choice in "${CHOICES[@]}"; do
        choice=$(echo "$choice" | tr -d ' ') # Remove spaces
        case "$choice" in
            1)
                if [[ " ${providers[*]} " =~ " github " ]]; then
                    jq --slurpfile github "$TEMP_DIR/github_repositories.json" \
                        '.repositories.github = $github[0]' \
                        "$TEMP_DIR/selected_repos.json" > "$TEMP_DIR/selected_repos.json.tmp"
                    mv "$TEMP_DIR/selected_repos.json.tmp" "$TEMP_DIR/selected_repos.json"
                    log_success "Added GitHub repositories"
                fi
                ;;
            2)
                if [[ " ${providers[*]} " =~ " gitlab " ]]; then
                    jq --slurpfile gitlab "$TEMP_DIR/gitlab_repositories.json" \
                        '.repositories.gitlab = $gitlab[0]' \
                        "$TEMP_DIR/selected_repos.json" > "$TEMP_DIR/selected_repos.json.tmp"
                    mv "$TEMP_DIR/selected_repos.json.tmp" "$TEMP_DIR/selected_repos.json"
                    log_success "Added GitLab projects"
                fi
                ;;
            3)
                if [[ " ${providers[*]} " =~ " bitbucket " ]]; then
                    jq --slurpfile bitbucket "$TEMP_DIR/bitbucket_repositories.json" \
                        '.repositories.bitbucket = $bitbucket[0]' \
                        "$TEMP_DIR/selected_repos.json" > "$TEMP_DIR/selected_repos.json.tmp"
                    mv "$TEMP_DIR/selected_repos.json.tmp" "$TEMP_DIR/selected_repos.json"
                    log_success "Added Bitbucket repositories"
                fi
                ;;
        esac
    done

    local total_selected
    total_selected=$(jq '.repositories.github + .repositories.gitlab + .repositories.bitbucket | length' "$TEMP_DIR/selected_repos.json")
    log_success "Selected $total_selected repositories from chosen providers"

    generate_configuration
}

# Interactive repository selection
select_repositories_interactively() {
    log_info "Interactive selection not yet implemented"
    log_info "For now, using 'select all' as fallback"
    select_all_repositories
}

# Configure merge strategies
configure_merge_strategies() {
    log_step "Configuring merge strategies..."
    echo ""
    echo "Available merge strategies:"
    echo "1. squash - Combine all commits into single commit (recommended)"
    echo "2. merge  - Preserve commit history with merge commit"
    echo "3. rebase - Linear history without merge commits"
    echo ""

    read -r -p "Choose default merge strategy (1-3): " strategy_choice

    local strategy
    case "$strategy_choice" in
        1) strategy="squash" ;;
        2) strategy="merge" ;;
        3) strategy="rebase" ;;
        *)
            log_warning "Invalid choice, using 'squash' as default"
            strategy="squash"
            ;;
    esac

    echo "$strategy" > "$TEMP_DIR/merge_strategy"
    log_success "Set default merge strategy to: $strategy"
}

# Preview configuration
preview_configuration() {
    log_step "Configuration Preview"

    if [[ ! -f "$TEMP_DIR/selected_repos.json" ]]; then
        log_warning "No repositories selected yet. Please select repositories first."
        return
    fi

    echo ""
    echo -e "${BOLD}${CYAN}Configuration Preview:${NC}"
    echo ""

    # Show selected repositories count
    local github_count gitlab_count bitbucket_count
    github_count=$(jq '.repositories.github | length' "$TEMP_DIR/selected_repos.json")
    gitlab_count=$(jq '.repositories.gitlab | length' "$TEMP_DIR/selected_repos.json")
    bitbucket_count=$(jq '.repositories.bitbucket | length' "$TEMP_DIR/selected_repos.json")

    echo -e "${GREEN}Selected Repositories:${NC}"
    echo "  GitHub: $github_count repositories"
    echo "  GitLab: $gitlab_count projects"
    echo "  Bitbucket: $bitbucket_count repositories"
    echo ""

    # Show merge strategy
    local merge_strategy
    if [[ -f "$TEMP_DIR/merge_strategy" ]]; then
        merge_strategy=$(cat "$TEMP_DIR/merge_strategy")
    else
        merge_strategy="squash"
    fi
    echo -e "${GREEN}Merge Strategy:${NC} $merge_strategy"
    echo ""

    # Show sample repositories
    if [[ "$github_count" -gt 0 ]]; then
        echo -e "${YELLOW}Sample GitHub repositories:${NC}"
        jq -r '.repositories.github[:3] | .[] | "  - " + .full_name' "$TEMP_DIR/selected_repos.json"
        if [[ "$github_count" -gt 3 ]]; then
            echo "  ... and $((github_count - 3)) more"
        fi
        echo ""
    fi

    if [[ "$gitlab_count" -gt 0 ]]; then
        echo -e "${YELLOW}Sample GitLab projects:${NC}"
        jq -r '.repositories.gitlab[:3] | .[] | "  - " + .path_with_namespace' "$TEMP_DIR/selected_repos.json"
        if [[ "$gitlab_count" -gt 3 ]]; then
            echo "  ... and $((gitlab_count - 3)) more"
        fi
        echo ""
    fi

    if [[ "$bitbucket_count" -gt 0 ]]; then
        echo -e "${YELLOW}Sample Bitbucket repositories:${NC}"
        jq -r '.repositories.bitbucket[:3] | .[] | "  - " + .full_name' "$TEMP_DIR/selected_repos.json"
        if [[ "$bitbucket_count" -gt 3 ]]; then
            echo "  ... and $((bitbucket_count - 3)) more"
        fi
        echo ""
    fi
}

# Generate final configuration
generate_configuration() {
    log_step "Generating configuration..."

    if [[ ! -f "$TEMP_DIR/selected_repos.json" ]]; then
        log_error "No repositories selected. Cannot generate configuration."
        return 1
    fi

    # Create the final configuration
    create_final_config

    # Validate the generated configuration
    if [[ "$PREVIEW_MODE" == false ]]; then
        validate_final_config "$CONFIG_FILE"
    fi

    # Show final summary
    preview_configuration

    echo ""
    if [[ "$PREVIEW_MODE" == false ]]; then
        log_success "Configuration generated successfully!"
        log_info "Configuration saved to: $CONFIG_FILE"

        # Show next steps
        echo ""
        echo -e "${BOLD}${YELLOW}Next Steps:${NC}"
        echo "1. Review your configuration: cat $CONFIG_FILE"
        echo "2. Validate configuration: make validate"
        echo "3. Test PR checking: make check-prs"
        echo "4. Test notification setup: make test-notifications"
        echo "5. Start monitoring: make watch"
    else
        log_info "Preview mode: Configuration was not saved"
        echo ""
        echo -e "${YELLOW}To generate the actual configuration:${NC}"
        echo "  ./setup-wizard.sh  # (without --preview flag)"
    fi
}

# Create the final configuration YAML file
create_final_config() {
    local final_config="$CONFIG_FILE"
    local temp_config="$TEMP_DIR/config_output.yaml"

    # Load settings
    local merge_strategy="squash"
    if [[ -f "$TEMP_DIR/merge_strategy" ]]; then
        merge_strategy=$(cat "$TEMP_DIR/merge_strategy")
    fi

    log_info "Creating configuration with merge strategy: $merge_strategy"

    # Start building the configuration
    cat > "$temp_config" << EOF
# Multi-Gitter Pull-Request Automation Configuration
# Generated by setup wizard on $(date)

# Global configuration
config:
  # Default merge strategy (squash, merge, rebase)
  default_merge_strategy: "$merge_strategy"
  # Auto-merge settings
  auto_merge:
    enabled: true
    # Wait for status checks to pass
    wait_for_checks: true
    # Require approval before merge
    require_approval: false
  # PR filters
  pr_filters:
    # Only process PRs from these actors
    allowed_actors:
      - "dependabot[bot]"
      - "renovate[bot]"
    # Skip PRs with these labels
    skip_labels:
      - "do-not-merge"
      - "wip"
      - "draft"

# Repository configurations
repositories:
EOF

    # Add GitHub repositories
    add_github_repos_to_config "$temp_config"

    # Add GitLab repositories
    add_gitlab_repos_to_config "$temp_config"

    # Add Bitbucket repositories
    add_bitbucket_repos_to_config "$temp_config"

    # Add authentication configuration
    cat >> "$temp_config" << EOF

# Authentication configuration (use environment variables for security)
auth:
  github:
    token: "\${GITHUB_TOKEN}"

  gitlab:
    token: "\${GITLAB_TOKEN}"
    url: "\${GITLAB_URL:-https://gitlab.com}"  # For self-hosted GitLab instances

  bitbucket:
    username: "\${BITBUCKET_USERNAME}"
    app_password: "\${BITBUCKET_APP_PASSWORD}"
    workspace: "\${BITBUCKET_WORKSPACE}"

# Notification settings (optional)
notifications:
  slack:
    webhook_url: "\${SLACK_WEBHOOK_URL}"
    channel: "#deployments"
    enabled: false

  email:
    smtp_server: "smtp.gmail.com"
    smtp_port: 587
    username: "\${EMAIL_USERNAME}"
    password: "\${EMAIL_PASSWORD}"
    recipient: "\${EMAIL_RECIPIENT}"  # Optional, defaults to username
    enabled: false
EOF

    # Handle existing config file
    if [[ "$ADDITIVE_MODE" == true ]] && [[ -f "$final_config" ]]; then
        log_info "Additive mode: Merging with existing configuration"
        merge_with_existing_config "$temp_config" "$final_config"
    else
        if [[ "$PREVIEW_MODE" == false ]]; then
            cp "$temp_config" "$final_config"
            log_success "Configuration written to $final_config"
        else
            log_info "Preview mode: Configuration not saved"
        fi
    fi
}

# Add GitHub repositories to configuration
add_github_repos_to_config() {
    local config_file="$1"
    local github_count
    github_count=$(jq '.repositories.github | length' "$TEMP_DIR/selected_repos.json")

    echo "  # GitHub repositories" >> "$config_file"
    echo "  github:" >> "$config_file"

    if [[ "$github_count" -gt 0 ]]; then
        jq -r '.repositories.github[] | "    - name: \"" + .full_name + "\"\n      url: \"" + .html_url + "\"\n      provider: \"github\"\n      auth_type: \"token\"\n      merge_strategy: \"squash\"\n      auto_merge: true\n"' \
            "$TEMP_DIR/selected_repos.json" >> "$config_file"
    else
        echo "    []" >> "$config_file"
    fi

    log_success "Added $github_count GitHub repositories to configuration"
}

# Add GitLab repositories to configuration
add_gitlab_repos_to_config() {
    local config_file="$1"
    local gitlab_count
    gitlab_count=$(jq '.repositories.gitlab | length' "$TEMP_DIR/selected_repos.json")

    echo "" >> "$config_file"
    echo "  # GitLab repositories" >> "$config_file"
    echo "  gitlab:" >> "$config_file"

    if [[ "$gitlab_count" -gt 0 ]]; then
        jq -r '.repositories.gitlab[] | "    - name: \"" + .path_with_namespace + "\"\n      url: \"" + .web_url + "\"\n      provider: \"gitlab\"\n      auth_type: \"token\"\n      merge_strategy: \"squash\"\n      auto_merge: true\n"' \
            "$TEMP_DIR/selected_repos.json" >> "$config_file"
    else
        echo "    []" >> "$config_file"
    fi

    log_success "Added $gitlab_count GitLab projects to configuration"
}

# Add Bitbucket repositories to configuration
add_bitbucket_repos_to_config() {
    local config_file="$1"
    local bitbucket_count
    bitbucket_count=$(jq '.repositories.bitbucket | length' "$TEMP_DIR/selected_repos.json")

    echo "" >> "$config_file"
    echo "  # Bitbucket repositories" >> "$config_file"
    echo "  bitbucket:" >> "$config_file"

    if [[ "$bitbucket_count" -gt 0 ]]; then
        jq -r '.repositories.bitbucket[] | "    - name: \"" + .full_name + "\"\n      url: \"" + .links.html.href + "\"\n      provider: \"bitbucket\"\n      auth_type: \"app-password\"\n      merge_strategy: \"squash\"\n      auto_merge: true\n"' \
            "$TEMP_DIR/selected_repos.json" >> "$config_file"
    else
        echo "    []" >> "$config_file"
    fi

    log_success "Added $bitbucket_count Bitbucket repositories to configuration"
}

# Merge with existing configuration (additive mode)
merge_with_existing_config() {
    local temp_config="$1"
    local final_config="$2"

    log_info "Merging configurations..."

    # Create a backup of the existing config
    local backup_file="${final_config}.backup.$(date +%Y%m%d_%H%M%S)"
    cp "$final_config" "$backup_file"
    log_info "Backed up existing config to: $backup_file"

    # For now, we'll use a simple approach: append new repositories to existing ones
    # A more sophisticated merge would use yq to properly merge YAML structures

    if [[ "$PREVIEW_MODE" == false ]]; then
        # Extract new repositories and append to existing config
        echo "" >> "$final_config"
        echo "# === Repositories added by setup wizard on $(date) ===" >> "$final_config"

        # Extract repository sections from temp config
        sed -n '/^repositories:/,/^auth:/p' "$temp_config" | sed '$d' >> "$final_config"

        log_success "Merged new repositories with existing configuration"
        log_info "Original config backed up to: $backup_file"
    fi
}

# Enhanced error handling and validation
validate_environment() {
    log_step "Validating environment..."

    local validation_errors=0

    # Check for required tools with specific version requirements
    if ! command -v jq &> /dev/null; then
        log_error "jq is required but not installed"
        ((validation_errors++))
    else
        local jq_version
        jq_version=$(jq --version 2>/dev/null | grep -o '[0-9]\+\.[0-9]\+' | head -1)
        if [[ -n "$jq_version" ]]; then
            log_info "jq version: $jq_version"
        fi
    fi

    if ! command -v yq &> /dev/null; then
        log_error "yq is required but not installed"
        ((validation_errors++))
    else
        local yq_version
        yq_version=$(yq --version 2>/dev/null | grep -o '[0-9]\+\.[0-9]\+\.[0-9]\+' | head -1)
        if [[ -n "$yq_version" ]]; then
            log_info "yq version: $yq_version"
        fi
    fi

    if ! command -v curl &> /dev/null; then
        log_error "curl is required but not installed"
        ((validation_errors++))
    fi

    # Check disk space for temp directory
    local available_space
    available_space=$(df /tmp | awk 'NR==2 {print $4}')
    if [[ "$available_space" -lt 10000 ]]; then
        log_warning "Low disk space in /tmp (${available_space}KB available)"
    fi

    # Validate network connectivity
    if ! curl -s --connect-timeout 5 https://api.github.com/user/repos >/dev/null 2>&1; then
        log_warning "Network connectivity check failed - some features may not work"
    fi

    if [[ "$validation_errors" -gt 0 ]]; then
        log_error "Environment validation failed with $validation_errors errors"
        log_info "Please run 'make install' to install missing dependencies"
        return 1
    fi

    log_success "Environment validation passed"
    return 0
}

# Enhanced API error handling
handle_api_error() {
    local provider="$1"
    local response="$2"
    local endpoint="$3"

    if [[ -z "$response" ]]; then
        log_error "$provider API: No response from $endpoint"
        return 1
    fi

    # Check for common API error patterns
    if echo "$response" | jq -e '.message' &>/dev/null; then
        local error_message
        error_message=$(echo "$response" | jq -r '.message')
        log_error "$provider API: $error_message"
        return 1
    fi

    # Check HTTP error responses
    if echo "$response" | grep -q '"Bad credentials"'; then
        log_error "$provider API: Authentication failed - check your token"
        return 1
    fi

    if echo "$response" | grep -q '"rate limit"'; then
        log_error "$provider API: Rate limit exceeded - please wait and try again"
        return 1
    fi

    return 0
}

# Validate configuration completeness
validate_final_config() {
    local config_file="$1"

    log_step "Validating generated configuration..."

    if [[ ! -f "$config_file" ]]; then
        log_error "Configuration file not found: $config_file"
        return 1
    fi

    # Validate YAML syntax
    if ! yq eval '.' "$config_file" >/dev/null 2>&1; then
        log_error "Invalid YAML syntax in configuration file"
        return 1
    fi

    # Check for required sections
    local required_sections=("config" "repositories" "auth")
    for section in "${required_sections[@]}"; do
        if ! yq eval ".$section" "$config_file" >/dev/null 2>&1; then
            log_error "Missing required section: $section"
            return 1
        fi
    done

    # Validate repository count
    local total_repos
    total_repos=$(yq eval '.repositories | (.github // []) + (.gitlab // []) + (.bitbucket // []) | length' "$config_file")

    if [[ "$total_repos" -eq 0 ]]; then
        log_warning "No repositories configured - this configuration won't do anything"
    else
        log_success "Configuration contains $total_repos repositories"
    fi

    # Check for environment variables
    local env_vars
    env_vars=$(grep -o '\${[^}]*}' "$config_file" | sort -u || true)

    if [[ -n "$env_vars" ]]; then
        echo ""
        echo -e "${YELLOW}Required environment variables:${NC}"
        echo "$env_vars" | while read -r var; do
            local var_name
            var_name=$(echo "$var" | sed 's/\${//' | sed 's/}//' | sed 's/:-.*//')
            if [[ -n "${!var_name:-}" ]]; then
                echo -e "  ${GREEN}✓${NC} $var_name (set)"
            else
                echo -e "  ${RED}✗${NC} $var_name (not set)"
            fi
        done
    fi

    log_success "Configuration validation completed"
    return 0
}

# Graceful cleanup on script exit
cleanup_on_exit() {
    local exit_code=$?

    if [[ "$exit_code" -ne 0 ]]; then
        log_error "Script exited with errors (code: $exit_code)"

        # Show helpful troubleshooting info
        echo ""
        echo -e "${YELLOW}Troubleshooting tips:${NC}"
        echo "1. Check your internet connection"
        echo "2. Verify your authentication tokens are valid"
        echo "3. Ensure you have sufficient API rate limits"
        echo "4. Check the logs above for specific error messages"
        echo ""
        echo -e "${BLUE}For help:${NC}"
        echo "- Run with --help for usage information"
        echo "- Check the README for setup instructions"
        echo "- Verify environment variables are set correctly"
    fi

    # Clean up temp directory
    if [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR" 2>/dev/null || true
    fi
}

# Progress tracking for long operations
show_progress() {
    local current="$1"
    local total="$2"
    local operation="$3"

    local percent=$((current * 100 / total))
    local bar_length=30
    local filled_length=$((percent * bar_length / 100))

    printf "\r${BLUE}[INFO]${NC} $operation: ["

    for ((i=0; i<filled_length; i++)); do
        printf "▓"
    done

    for ((i=filled_length; i<bar_length; i++)); do
        printf "░"
    done

    printf "] %d%% (%d/%d)" "$percent" "$current" "$total"

    if [[ "$current" -eq "$total" ]]; then
        echo ""  # New line when complete
    fi
}

# Enhanced input validation
validate_input() {
    local input="$1"
    local type="$2"
    local pattern="$3"

    case "$type" in
        "number")
            if [[ ! "$input" =~ ^[0-9]+$ ]]; then
                return 1
            fi
            ;;
        "choice")
            if [[ ! "$input" =~ $pattern ]]; then
                return 1
            fi
            ;;
        "non-empty")
            if [[ -z "$input" ]]; then
                return 1
            fi
            ;;
    esac

    return 0
}

# Main execution function
main() {
    # Set up cleanup trap
    trap cleanup_on_exit EXIT

    # Parse arguments first (this also creates temp dir)
    parse_arguments "$@"

    # Show welcome message
    show_welcome

    # Enhanced environment validation
    validate_environment || exit 1

    # Backup config (temp dir already created)
    backup_config

    # Check authentication with enhanced error handling
    check_authentication || exit 1

    log_header "Configuration Wizard Ready"
    log_success "All checks passed - ready to discover repositories!"

    # Start the repository discovery process
    discover_repositories

    if [[ "$PREVIEW_MODE" == true ]]; then
        log_info "Running in preview mode - no changes will be made"
    fi

    if [[ "$ADDITIVE_MODE" == true ]]; then
        log_info "Running in additive mode - will preserve existing configuration"
    fi
}

# Execute main function with all arguments
main "$@"