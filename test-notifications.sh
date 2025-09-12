#!/bin/bash

# test-notifications.sh - Test notification functionality
# This script tests Slack and email notifications

set -euo pipefail

CONFIG_FILE="${CONFIG_FILE:-config.yaml}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# Test Slack notification
test_slack_notification() {
    log_info "Testing Slack notification..."

    if [[ -z "${SLACK_WEBHOOK_URL:-}" ]]; then
        log_warning "SLACK_WEBHOOK_URL not set, skipping Slack test"
        return 0
    fi

    local slack_message="🧪 Test notification from multi-gitter automation"

    if curl -s -X POST \
        -H 'Content-type: application/json' \
        --data "{\"text\":\"$slack_message\"}" \
        "$SLACK_WEBHOOK_URL" > /dev/null; then
        log_success "Slack notification sent successfully"
    else
        log_error "Failed to send Slack notification"
        return 1
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
        log_success "Email notification sent successfully"
        rm -f "$email_file"
        return 0
    else
        log_error "Failed to send email notification"
        rm -f "$email_file"
        return 1
    fi
}

# Test email notification
test_email_notification() {
    log_info "Testing email notification..."

    if [[ -z "${EMAIL_USERNAME:-}" ]] || [[ -z "${EMAIL_PASSWORD:-}" ]]; then
        log_warning "EMAIL_USERNAME or EMAIL_PASSWORD not set, skipping email test"
        return 0
    fi

    local smtp_server smtp_port email_subject email_body
    smtp_server=$(yq '.notifications.email.smtp_server' "$CONFIG_FILE" 2>/dev/null || echo "smtp.gmail.com")
    smtp_port=$(yq '.notifications.email.smtp_port' "$CONFIG_FILE" 2>/dev/null || echo "587")

    email_subject="🧪 Test notification from multi-gitter automation"
    email_body="This is a test email notification from multi-gitter automation.

If you received this email, your email notification configuration is working correctly.

Configuration:
- SMTP Server: $smtp_server
- SMTP Port: $smtp_port
- Username: $EMAIL_USERNAME

This is an automated test notification."

    send_email "$EMAIL_USERNAME" "$EMAIL_PASSWORD" "$smtp_server" "$smtp_port" "$email_subject" "$email_body"
}

# Test notification configurations
test_notification_config() {
    log_info "Testing notification configuration..."

    if [[ ! -f "$CONFIG_FILE" ]]; then
        log_error "Configuration file not found: $CONFIG_FILE"
        return 1
    fi

    # Check Slack config
    local slack_enabled
    slack_enabled=$(yq '.notifications.slack.enabled' "$CONFIG_FILE" 2>/dev/null || echo "false")
    log_info "Slack notifications enabled: $slack_enabled"

    if [[ "$slack_enabled" == "true" ]]; then
        if [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
            log_success "Slack webhook URL configured"
        else
            log_warning "Slack enabled but SLACK_WEBHOOK_URL not set"
        fi
    fi

    # Check email config
    local email_enabled
    email_enabled=$(yq '.notifications.email.enabled' "$CONFIG_FILE" 2>/dev/null || echo "false")
    log_info "Email notifications enabled: $email_enabled"

    if [[ "$email_enabled" == "true" ]]; then
        if [[ -n "${EMAIL_USERNAME:-}" ]] && [[ -n "${EMAIL_PASSWORD:-}" ]]; then
            log_success "Email credentials configured"
        else
            log_warning "Email enabled but EMAIL_USERNAME or EMAIL_PASSWORD not set"
        fi
    fi
}

# Show usage information
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Test notification functionality for multi-gitter automation

OPTIONS:
    -h, --help          Show this help message
    -c, --config FILE   Configuration file (default: config.yaml)
    -s, --slack         Test only Slack notifications
    -e, --email         Test only email notifications
    --config-test       Test notification configuration only

ENVIRONMENT VARIABLES:
    SLACK_WEBHOOK_URL   Slack webhook URL
    EMAIL_USERNAME      Email username for SMTP
    EMAIL_PASSWORD      Email password for SMTP

EXAMPLES:
    $0                          # Test all configured notifications
    $0 --slack                  # Test only Slack notifications
    $0 --email                  # Test only email notifications
    $0 --config-test            # Test configuration only

EOF
}

# Main function
main() {
    local test_slack=true
    local test_email=true
    local config_only=false

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
            -s|--slack)
                test_slack=true
                test_email=false
                shift
                ;;
            -e|--email)
                test_slack=false
                test_email=true
                shift
                ;;
            --config-test)
                config_only=true
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done

    log_info "Starting notification tests..."

    # Always test configuration
    test_notification_config

    if [[ "$config_only" == "true" ]]; then
        log_info "Configuration test completed"
        return 0
    fi

    # Test notifications
    local test_count=0
    local passed_count=0

    if [[ "$test_slack" == "true" ]]; then
        ((test_count++))
        if test_slack_notification; then
            ((passed_count++))
        fi
    fi

    if [[ "$test_email" == "true" ]]; then
        ((test_count++))
        if test_email_notification; then
            ((passed_count++))
        fi
    fi

    log_info "Tests completed: $passed_count/$test_count passed"

    if [[ $passed_count -eq $test_count ]]; then
        log_success "All notification tests passed!"
        return 0
    else
        log_warning "Some notification tests failed"
        return 1
    fi
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
            echo "brew install yq"
            ;;
        "linux")
            if command -v apt-get &> /dev/null; then
                echo "sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 && sudo chmod +x /usr/local/bin/yq"
            elif command -v yum &> /dev/null; then
                echo "sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 && sudo chmod +x /usr/local/bin/yq"
            elif command -v dnf &> /dev/null; then
                echo "sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 && sudo chmod +x /usr/local/bin/yq"
            elif command -v pacman &> /dev/null; then
                echo "sudo pacman -S yq"
            else
                echo "Install yq using your package manager, or see: https://github.com/mikefarah/yq#install"
            fi
            ;;
        *)
            echo "Install yq using your package manager"
            ;;
    esac
}

# Check dependencies
if ! command -v yq &> /dev/null; then
    log_error "yq is required but not installed. Run: $(get_install_command)"
    exit 1
fi

if ! command -v curl &> /dev/null; then
    log_error "curl is required but not installed (usually pre-installed on most systems)"
    exit 1
fi

# Run main function
main "$@"