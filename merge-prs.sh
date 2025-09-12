#!/bin/bash

# merge-prs.sh - Approve and merge ready PRs across multiple repositories
# This script reads the config.yaml file and merges PRs that are ready

set -euo pipefail

# Configuration
CONFIG_FILE="${CONFIG_FILE:-config.yaml}"
DRY_RUN="${DRY_RUN:-false}"
FORCE="${FORCE:-false}"

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

# Get merge strategy for repository
get_merge_strategy() {
    local provider="$1"
    local repo_name="$2"

    # Try to get repo-specific strategy first
    local strategy
    strategy=$(yq ".repositories.$provider[] | select(.name == \"$repo_name\") | .merge_strategy" "$CONFIG_FILE" 2>/dev/null || echo "")

    # Fall back to default strategy
    if [[ -z "$strategy" ]] || [[ "$strategy" == "null" ]]; then
        strategy=$(yq '.config.default_merge_strategy' "$CONFIG_FILE" 2>/dev/null || echo "squash")
    fi

    echo "$strategy"
}

# Check if auto-merge is enabled for repository
is_auto_merge_enabled() {
    local provider="$1"
    local repo_name="$2"

    # Try to get repo-specific setting first
    local auto_merge
    auto_merge=$(yq ".repositories.$provider[] | select(.name == \"$repo_name\") | .auto_merge" "$CONFIG_FILE" 2>/dev/null || echo "")

    # Fall back to global setting
    if [[ -z "$auto_merge" ]] || [[ "$auto_merge" == "null" ]]; then
        auto_merge=$(yq '.config.auto_merge.enabled' "$CONFIG_FILE" 2>/dev/null || echo "true")
    fi

    [[ "$auto_merge" == "true" ]]
}

# Merge GitHub PR
merge_github_pr() {
    local repo_name="$1"
    local pr_number="$2"
    local merge_strategy="$3"

    if [[ -z "$GITHUB_TOKEN" ]]; then
        log_error "GITHUB_TOKEN environment variable not set"
        return 1
    fi

    log_info "Processing GitHub PR: $repo_name#$pr_number"

    # Get PR details
    local pr_json
    pr_json=$(gh api "repos/$repo_name/pulls/$pr_number" 2>/dev/null || echo "{}")

    if [[ "$pr_json" == "{}" ]]; then
        log_error "Could not fetch PR details for $repo_name#$pr_number"
        return 1
    fi

    local pr_title pr_author pr_mergeable pr_checks
    pr_title=$(echo "$pr_json" | jq -r '.title')
    pr_author=$(echo "$pr_json" | jq -r '.user.login')
    pr_mergeable=$(echo "$pr_json" | jq -r '.mergeable')

    # Check allowed actors
    local allowed_actors
    allowed_actors=$(yq '.config.pr_filters.allowed_actors[]' "$CONFIG_FILE" 2>/dev/null || echo "")

    if [[ -n "$allowed_actors" ]]; then
        local author_allowed=false
        while IFS= read -r allowed_actor; do
            if [[ "$pr_author" == "$allowed_actor" ]]; then
                author_allowed=true
                break
            fi
        done <<< "$allowed_actors"

        if [[ "$author_allowed" == false ]]; then
            log_warning "Author $pr_author not in allowed actors list, skipping PR $repo_name#$pr_number"
            return 2
        fi
    fi

    # Check status checks
    local require_checks
    require_checks=$(yq '.config.auto_merge.wait_for_checks' "$CONFIG_FILE" 2>/dev/null || echo "true")

    if [[ "$require_checks" == "true" ]]; then
        local pr_sha
        pr_sha=$(echo "$pr_json" | jq -r '.head.sha')

        # Try combined status first (covers most check types)
        local status_response
        status_response=$(gh api "repos/$repo_name/commits/$pr_sha/status" 2>/dev/null || echo '{"state": "unknown"}')

        local combined_state
        combined_state=$(echo "$status_response" | jq -r '.state')

        case "$combined_state" in
            "failure"|"error")
                log_warning "Status checks failing for PR $repo_name#$pr_number, skipping"
                return 2
                ;;
            "pending")
                log_warning "Status checks still pending for PR $repo_name#$pr_number, skipping"
                return 2
                ;;
            "success")
                # All checks passing, continue
                ;;
            *)
                # Fallback to check-runs API if combined status is unavailable
                local checks_response
                checks_response=$(gh api "repos/$repo_name/commits/$pr_sha/check-runs" 2>/dev/null || echo '{"check_runs": []}')

                if [[ "$checks_response" != '{"check_runs": []}' ]]; then
                    local total_checks success_checks
                    total_checks=$(echo "$checks_response" | jq '.total_count // (.check_runs | length)')
                    success_checks=$(echo "$checks_response" | jq '[.check_runs[] | select(.conclusion == "success")] | length')

                    if [[ "$total_checks" -gt 0 ]] && [[ "$success_checks" -ne "$total_checks" ]]; then
                        log_warning "Status checks not passing for PR $repo_name#$pr_number, skipping"
                        return 2
                    fi
                fi
                # If no checks found or all successful, continue
                ;;
        esac
    fi

    # Check if PR is mergeable
    if [[ "$pr_mergeable" != "true" ]] && [[ "$pr_mergeable" != "null" ]] && [[ "$FORCE" != "true" ]]; then
        log_warning "PR $repo_name#$pr_number is not mergeable, skipping"
        return 2
    fi

    # Check if approval is required
    local require_approval
    require_approval=$(yq '.config.auto_merge.require_approval' "$CONFIG_FILE" 2>/dev/null || echo "true")

    if [[ "$require_approval" == "true" ]]; then
        # Check if PR is already approved
        local reviews_json
        reviews_json=$(gh api "repos/$repo_name/pulls/$pr_number/reviews" 2>/dev/null || echo "[]")

        local approved=false
        if [[ "$reviews_json" != "[]" ]]; then
            local approved_reviews
            approved_reviews=$(echo "$reviews_json" | jq '[.[] | select(.state == "APPROVED")] | length')
            [[ "$approved_reviews" -gt 0 ]] && approved=true
        fi

        # Approve PR if not already approved
        if [[ "$approved" == false ]]; then
            if [[ "$DRY_RUN" == "true" ]]; then
                log_info "[DRY RUN] Would approve PR $repo_name#$pr_number"
            else
                log_info "Approving PR $repo_name#$pr_number"
                if gh pr review "$pr_number" --repo "$repo_name" --approve --body "Auto-approved by multi-gitter automation" 2>/dev/null; then
                    log_success "Approved PR $repo_name#$pr_number"
                else
                    log_error "Failed to approve PR $repo_name#$pr_number"
                    return 1
                fi
            fi
        else
            log_info "PR $repo_name#$pr_number is already approved"
        fi
    fi

    # Convert merge strategy to GitHub CLI format
    local gh_merge_flag
    case "$merge_strategy" in
        "squash")
            gh_merge_flag="--squash"
            ;;
        "merge")
            gh_merge_flag="--merge"
            ;;
        "rebase")
            gh_merge_flag="--rebase"
            ;;
        *)
            log_warning "Unknown merge strategy: $merge_strategy, defaulting to squash"
            gh_merge_flag="--squash"
            ;;
    esac

    # Merge PR
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would merge PR $repo_name#$pr_number using $merge_strategy strategy"
    else
        log_info "Merging PR $repo_name#$pr_number using $merge_strategy strategy"
        if gh pr merge "$pr_number" --repo "$repo_name" "$gh_merge_flag" --auto 2>/dev/null; then
            log_success "Successfully merged PR $repo_name#$pr_number"

            # Send notification if configured
            send_notification "GitHub" "$repo_name" "$pr_number" "$pr_title" "$pr_author" "merged"
        else
            log_error "Failed to merge PR $repo_name#$pr_number"
            return 1
        fi
    fi
}

# Merge GitLab MR
merge_gitlab_mr() {
    local repo_name="$1"
    local mr_iid="$2"
    local merge_strategy="$3"

    if [[ -z "$GITLAB_TOKEN" ]]; then
        log_error "GITLAB_TOKEN environment variable not set"
        return 1
    fi

    log_info "Processing GitLab MR: $repo_name!$mr_iid"

    local project_id
    project_id=$(echo "$repo_name" | sed 's/\//%2F/g')

    local gitlab_url
    gitlab_url=$(yq '.auth.gitlab.url' "$CONFIG_FILE" 2>/dev/null || echo "https://gitlab.com")

    # Get MR details
    local mr_json
    mr_json=$(curl -s --header "PRIVATE-TOKEN: $GITLAB_TOKEN" \
        "$gitlab_url/api/v4/projects/$project_id/merge_requests/$mr_iid" || echo "{}")

    if [[ "$mr_json" == "{}" ]]; then
        log_error "Could not fetch MR details for $repo_name!$mr_iid"
        return 1
    fi

    local mr_title mr_author mr_state mr_mergeable
    mr_title=$(echo "$mr_json" | jq -r '.title')
    mr_author=$(echo "$mr_json" | jq -r '.author.username')
    mr_state=$(echo "$mr_json" | jq -r '.state')
    mr_mergeable=$(echo "$mr_json" | jq -r '.merge_status')

    # Check if MR is in correct state
    if [[ "$mr_state" != "opened" ]]; then
        log_warning "MR $repo_name!$mr_iid is not in opened state, skipping"
        return 2
    fi

    if [[ "$mr_mergeable" != "can_be_merged" ]] && [[ "$FORCE" != "true" ]]; then
        log_warning "MR $repo_name!$mr_iid is not mergeable, skipping"
        return 2
    fi

    # Convert merge strategy to GitLab API format
    local merge_commit_message squash_commit_message
    merge_commit_message="Merge branch '$(echo "$mr_json" | jq -r '.source_branch')' into '$(echo "$mr_json" | jq -r '.target_branch')'"
    squash_commit_message="$mr_title"

    local merge_params
    case "$merge_strategy" in
        "squash")
            merge_params="squash=true&squash_commit_message=$(printf "%s" "$squash_commit_message" | jq -sRr @uri)"
            ;;
        "merge")
            merge_params="merge_commit_message=$(printf "%s" "$merge_commit_message" | jq -sRr @uri)"
            ;;
        *)
            log_warning "Unknown merge strategy for GitLab: $merge_strategy, defaulting to merge"
            merge_params="merge_commit_message=$(printf "%s" "$merge_commit_message" | jq -sRr @uri)"
            ;;
    esac

    # Merge MR
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would merge MR $repo_name!$mr_iid using $merge_strategy strategy"
    else
        log_info "Merging MR $repo_name!$mr_iid using $merge_strategy strategy"

        local response
        response=$(curl -s -X PUT \
            --header "PRIVATE-TOKEN: $GITLAB_TOKEN" \
            --header "Content-Type: application/json" \
            "$gitlab_url/api/v4/projects/$project_id/merge_requests/$mr_iid/merge?$merge_params")

        local merge_status
        merge_status=$(echo "$response" | jq -r '.state // "error"')

        if [[ "$merge_status" == "merged" ]]; then
            log_success "Successfully merged MR $repo_name!$mr_iid"

            # Send notification if configured
            send_notification "GitLab" "$repo_name" "$mr_iid" "$mr_title" "$mr_author" "merged"
        else
            local error_message
            error_message=$(echo "$response" | jq -r '.message // "Unknown error"')
            log_error "Failed to merge MR $repo_name!$mr_iid: $error_message"
            return 1
        fi
    fi
}

# Merge Bitbucket PR
merge_bitbucket_pr() {
    local repo_name="$1"
    local pr_id="$2"
    local merge_strategy="$3"

    if [[ -z "$BITBUCKET_USERNAME" ]] || [[ -z "$BITBUCKET_APP_PASSWORD" ]]; then
        log_error "BITBUCKET_USERNAME or BITBUCKET_APP_PASSWORD environment variable not set"
        return 1
    fi

    log_info "Processing Bitbucket PR: $repo_name#$pr_id"

    # Get PR details
    local pr_json
    pr_json=$(curl -s -u "$BITBUCKET_USERNAME:$BITBUCKET_APP_PASSWORD" \
        "https://api.bitbucket.org/2.0/repositories/$repo_name/pullrequests/$pr_id" || echo "{}")

    if [[ "$pr_json" == "{}" ]]; then
        log_error "Could not fetch PR details for $repo_name#$pr_id"
        return 1
    fi

    local pr_title pr_author pr_state
    pr_title=$(echo "$pr_json" | jq -r '.title')
    pr_author=$(echo "$pr_json" | jq -r '.author.username')
    pr_state=$(echo "$pr_json" | jq -r '.state')

    # Check if PR is in correct state
    if [[ "$pr_state" != "OPEN" ]]; then
        log_warning "PR $repo_name#$pr_id is not in OPEN state, skipping"
        return 2
    fi

    # Convert merge strategy to Bitbucket API format
    local merge_type
    case "$merge_strategy" in
        "squash")
            merge_type="squash"
            ;;
        "merge")
            merge_type="merge_commit"
            ;;
        *)
            log_warning "Unknown merge strategy for Bitbucket: $merge_strategy, defaulting to merge_commit"
            merge_type="merge_commit"
            ;;
    esac

    # Merge PR
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would merge PR $repo_name#$pr_id using $merge_strategy strategy"
    else
        log_info "Merging PR $repo_name#$pr_id using $merge_strategy strategy"

        local response
        response=$(curl -s -X POST \
            -u "$BITBUCKET_USERNAME:$BITBUCKET_APP_PASSWORD" \
            -H "Content-Type: application/json" \
            -d "{\"type\": \"$merge_type\", \"message\": \"$pr_title\"}" \
            "https://api.bitbucket.org/2.0/repositories/$repo_name/pullrequests/$pr_id/merge")

        local merge_status
        merge_status=$(echo "$response" | jq -r '.state // "error"')

        if [[ "$merge_status" == "MERGED" ]]; then
            log_success "Successfully merged PR $repo_name#$pr_id"

            # Send notification if configured
            send_notification "Bitbucket" "$repo_name" "$pr_id" "$pr_title" "$pr_author" "merged"
        else
            local error_message
            error_message=$(echo "$response" | jq -r '.error.message // "Unknown error"')
            log_error "Failed to merge PR $repo_name#$pr_id: $error_message"
            return 1
        fi
    fi
}

# Send notification
send_notification() {
    local provider="$1"
    local repo="$2"
    local pr_id="$3"
    local title="$4"
    local author="$5"
    local action="$6"

    # Slack notification
    local slack_enabled
    slack_enabled=$(yq '.notifications.slack.enabled' "$CONFIG_FILE" 2>/dev/null || echo "false")

    if [[ "$slack_enabled" == "true" ]] && [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
        local slack_message
        slack_message="✅ *$action* $provider PR: \`$repo#$pr_id\` - $title (by @$author)"

        curl -s -X POST \
            -H 'Content-type: application/json' \
            --data "{\"text\":\"$slack_message\"}" \
            "$SLACK_WEBHOOK_URL" > /dev/null
    fi

    # Email notification
    local email_enabled
    email_enabled=$(yq '.notifications.email.enabled' "$CONFIG_FILE" 2>/dev/null || echo "false")

    if [[ "$email_enabled" == "true" ]] && [[ -n "${EMAIL_USERNAME:-}" ]] && [[ -n "${EMAIL_PASSWORD:-}" ]]; then
        local smtp_server smtp_port email_subject email_body
        smtp_server=$(yq '.notifications.email.smtp_server' "$CONFIG_FILE" 2>/dev/null || echo "smtp.gmail.com")
        smtp_port=$(yq '.notifications.email.smtp_port' "$CONFIG_FILE" 2>/dev/null || echo "587")

        email_subject="PR $action: $repo#$pr_id"
        email_body="$provider pull request has been $action:

Repository: $repo
PR/MR: #$pr_id
Title: $title
Author: $author
Action: $action

This is an automated notification from multi-gitter automation."

        # Send email using curl and SMTP
        send_email "$EMAIL_USERNAME" "$EMAIL_PASSWORD" "$smtp_server" "$smtp_port" "$email_subject" "$email_body"
    fi
}

# Send email notification via SMTP
send_email() {
    local username="$1"
    local password="$2"
    local smtp_server="$3"
    local smtp_port="$4"
    local subject="$5"
    local body="$6"

    # Get recipient email (use sender as recipient by default)
    local recipient
    recipient=$(yq '.notifications.email.recipient' "$CONFIG_FILE" 2>/dev/null || echo "$username")

    # Create temporary file for email content
    local email_file
    email_file=$(mktemp)

    # Build email headers and body
    cat > "$email_file" << EOF
From: $username
To: $recipient
Subject: $subject

$body
EOF

    # Send email using curl
    if curl --silent --url "smtps://$smtp_server:$smtp_port" \
        --ssl-reqd \
        --mail-from "$username" \
        --mail-rcpt "$recipient" \
        --upload-file "$email_file" \
        --user "$username:$password" 2>/dev/null; then
        log_info "Email notification sent successfully"
    else
        log_warning "Failed to send email notification"
    fi

    # Clean up
    rm -f "$email_file"
}

# Get ready PRs for merging
get_ready_prs() {
    local output
    output=$(OUTPUT_FORMAT=json ./check-prs.sh 2>/dev/null | jq -r 'select(.ready_to_merge == true) | "\(.repository):\(.pr_number)"')
    echo "$output"
}

# Main function
main() {
    log_info "Starting PR merge process across multiple repositories"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "Running in DRY RUN mode - no actual merges will be performed"
    fi

    check_dependencies
    parse_config

    # Get all ready PRs
    local ready_prs
    ready_prs=$(get_ready_prs)

    if [[ -z "$ready_prs" ]]; then
        log_info "No PRs ready for merging found"
        return 0
    fi

    local merge_count=0

    # Process each ready PR
    while IFS=':' read -r repo_name pr_id; do
        [[ -z "$repo_name" ]] && continue

        # Determine provider based on repository configuration
        local provider=""

        if yq ".repositories.github[] | select(.name == \"$repo_name\")" "$CONFIG_FILE" &> /dev/null; then
            provider="github"
        elif yq ".repositories.gitlab[] | select(.name == \"$repo_name\")" "$CONFIG_FILE" &> /dev/null; then
            provider="gitlab"
        elif yq ".repositories.bitbucket[] | select(.name == \"$repo_name\")" "$CONFIG_FILE" &> /dev/null; then
            provider="bitbucket"
        else
            log_error "Could not determine provider for repository: $repo_name"
            continue
        fi

        # Check if auto-merge is enabled for this repository
        if ! is_auto_merge_enabled "$provider" "$repo_name"; then
            log_info "Auto-merge disabled for $repo_name, skipping"
            continue
        fi

        # Get merge strategy
        local merge_strategy
        merge_strategy=$(get_merge_strategy "$provider" "$repo_name")

        # Merge based on provider
        case "$provider" in
            "github")
                if merge_github_pr "$repo_name" "$pr_id" "$merge_strategy"; then
                    ((merge_count++))
                fi
                ;;
            "gitlab")
                if merge_gitlab_mr "$repo_name" "$pr_id" "$merge_strategy"; then
                    ((merge_count++))
                fi
                ;;
            "bitbucket")
                if merge_bitbucket_pr "$repo_name" "$pr_id" "$merge_strategy"; then
                    ((merge_count++))
                fi
                ;;
        esac

        # Add delay between merges to avoid rate limiting
        sleep 2

    done <<< "$ready_prs"

    log_success "PR merge process completed. Merged $merge_count PRs"
}

# Show usage information
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Approve and merge ready PRs across multiple repositories defined in config.yaml

OPTIONS:
    -h, --help          Show this help message
    -c, --config FILE   Configuration file (default: config.yaml)
    -n, --dry-run       Show what would be done without actually doing it
    -f, --force         Force merge even if checks indicate PR is not mergeable

ENVIRONMENT VARIABLES:
    GITHUB_TOKEN        GitHub personal access token
    GITLAB_TOKEN        GitLab personal access token
    BITBUCKET_USERNAME  Bitbucket username
    BITBUCKET_APP_PASSWORD Bitbucket app password
    SLACK_WEBHOOK_URL   Slack webhook URL for notifications

EXAMPLES:
    $0                                  # Merge ready PRs with default settings
    $0 --dry-run                       # Show what would be merged without doing it
    $0 -c custom-config.yaml           # Use custom configuration file
    DRY_RUN=true $0                    # Set dry run mode via environment

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
        -n|--dry-run)
            DRY_RUN="true"
            shift
            ;;
        -f|--force)
            FORCE="true"
            shift
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