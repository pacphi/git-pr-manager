# Notifications Guide

Setup Slack and email notifications for PR automation events.

## Table of Contents

- [Slack Notifications](#slack-notifications)
- [Email Notifications](#email-notifications)
- [Testing Notifications](#testing-notifications)
- [Notification Events](#notification-events)
- [Troubleshooting](#troubleshooting)

## Slack Notifications

### Setup Slack Webhook

1. **Create a Slack App**:

   - Go to https://api.slack.com/apps
   - Click "Create New App" > "From scratch"
   - Name your app (e.g., "Multi-Gitter Bot")
   - Select your workspace

2. **Enable Incoming Webhooks**:
   - Go to "Incoming Webhooks" in your app settings
   - Toggle "Activate Incoming Webhooks" to On
   - Click "Add New Webhook to Workspace"
   - Select the channel to post to
   - Copy the webhook URL

3. **Configure Environment Variable**:

   ```bash
   export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
   ```

4. **Enable in Configuration**:

   ```yaml
   notifications:
     slack:
       webhook_url: "${SLACK_WEBHOOK_URL}"
       channel: "#deployments"
       enabled: true
   ```

### Slack Message Format

Notifications will appear like:

```text
✅ merged GitHub PR: `owner/repo#123` - Fix critical bug in authentication (by @dependabot[bot])
```

### Customizing Slack Messages

To customize messages, edit the `send_notification` function in `merge-prs.sh`:

```bash
# Current format
slack_message="✅ *$action* $provider PR: \`$repo#$pr_id\` - $title (by @$author)"

# Custom format examples
slack_message="🚀 *$action* PR in $repo: <https://github.com/$repo/pull/$pr_id|#$pr_id> - $title"
slack_message="✅ Automatically $action $provider PR #$pr_id in $repo by $author"
```

## Email Notifications

### SMTP Configuration

Configure SMTP settings in your config file:

```yaml
notifications:
  email:
    smtp_server: "smtp.gmail.com"      # SMTP server
    smtp_port: 587                     # SMTP port (587 for STARTTLS, 465 for SSL)
    username: "${EMAIL_USERNAME}"      # Your email address
    password: "${EMAIL_PASSWORD}"      # App password or regular password
    recipient: "${EMAIL_RECIPIENT}"    # Optional, defaults to username
    enabled: true
```

### Environment Variables

```bash
export EMAIL_USERNAME="your-email@gmail.com"
export EMAIL_PASSWORD="your-app-password"
export EMAIL_RECIPIENT="notifications@company.com"  # Optional
```

### Gmail Setup

1. **Enable 2-Factor Authentication** (if not already enabled)
2. **Create App Password**:
   - Go to Google Account settings
   - Security > App passwords
   - Select "Mail" and your device
   - Copy the generated password

3. **Use App Password**:

   ```bash
   export EMAIL_PASSWORD="abcd efgh ijkl mnop"  # 16-character app password
   ```

### Other Email Providers

#### Outlook/Hotmail

```yaml
email:
  smtp_server: "smtp-mail.outlook.com"
  smtp_port: 587
```

#### Yahoo Mail

```yaml
email:
  smtp_server: "smtp.mail.yahoo.com"
  smtp_port: 587
```

#### Custom SMTP

```yaml
email:
  smtp_server: "mail.yourcompany.com"
  smtp_port: 587  # or 465 for SSL
```

### Email Message Format

Notifications will be sent with:

**Subject**: `PR merged: owner/repo#123`

**Body**:

```text
GitHub pull request has been merged:

Repository: owner/repo
PR/MR: #123
Title: Fix critical bug in authentication
Author: dependabot[bot]
Action: merged

This is an automated notification from multi-gitter automation.
```

## Testing Notifications

### Test All Notifications

```bash
./test-notifications.sh
```

### Test Specific Notification Types

```bash
# Test only Slack
./test-notifications.sh --slack

# Test only email
./test-notifications.sh --email

# Test configuration only
./test-notifications.sh --config-test
```

### Manual Testing

#### Test Slack Manually

```bash
curl -X POST \
  -H 'Content-type: application/json' \
  --data '{"text":"🧪 Test message from multi-gitter automation"}' \
  "$SLACK_WEBHOOK_URL"
```

#### Test Email Manually

```bash
# Create test email
cat > test-email.txt << EOF
From: $EMAIL_USERNAME
To: $EMAIL_USERNAME
Subject: Test Email

This is a test email from multi-gitter automation.
EOF

# Send via curl
curl --url "smtps://smtp.gmail.com:587" \
  --ssl-reqd \
  --mail-from "$EMAIL_USERNAME" \
  --mail-rcpt "$EMAIL_USERNAME" \
  --upload-file test-email.txt \
  --user "$EMAIL_USERNAME:$EMAIL_PASSWORD"
```

## Notification Events

Notifications are triggered when:

### Successful PR Merge

- **Trigger**: PR/MR successfully merged
- **Message**: Includes repository, PR number, title, and author
- **Icon**: ✅

### Future Events (Extensible)

The notification system can be extended for:

- **PR Approval**: When PR is approved but not yet merged
- **Merge Failures**: When PR merge attempts fail
- **Status Check Updates**: When CI/CD status changes
- **New PRs**: When new PRs are opened by watched actors

Example extension in `check-prs.sh`:

```bash
# Notify when new PR is found
if [[ "$pr_created_at" > "$last_check_time" ]]; then
    send_notification "GitHub" "$repo_name" "$pr_number" "$pr_title" "$pr_author" "opened"
fi
```

## Advanced Configuration

### Multiple Slack Channels

Send notifications to different channels based on repository:

```bash
# In send_notification function
case "$repo" in
    "company/frontend-*")
        SLACK_WEBHOOK_URL="$FRONTEND_SLACK_WEBHOOK"
        ;;
    "company/backend-*")
        SLACK_WEBHOOK_URL="$BACKEND_SLACK_WEBHOOK"
        ;;
    *)
        SLACK_WEBHOOK_URL="$DEFAULT_SLACK_WEBHOOK"
        ;;
esac
```

### Conditional Notifications

Only notify for specific events:

```yaml
notifications:
  slack:
    enabled: true
    conditions:
      - provider: "github"
        authors: ["dependabot[bot]", "renovate[bot]"]
      - provider: "gitlab"
        projects: ["critical/*"]
```

### Rich Slack Messages

Use Slack's block kit for rich formatting:

```bash
slack_payload=$(cat << EOF
{
  "blocks": [
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "✅ *PR Merged* in \`$repo\`"
      }
    },
    {
      "type": "section",
      "fields": [
        {
          "type": "mrkdwn",
          "text": "*PR:* <https://github.com/$repo/pull/$pr_id|#$pr_id>"
        },
        {
          "type": "mrkdwn",
          "text": "*Author:* $author"
        },
        {
          "type": "mrkdwn",
          "text": "*Title:* $title"
        }
      ]
    }
  ]
}
EOF
)

curl -X POST \
  -H 'Content-type: application/json' \
  --data "$slack_payload" \
  "$SLACK_WEBHOOK_URL"
```

## Troubleshooting

### Slack Issues

#### Webhook Not Working

1. **Verify webhook URL** is correct and active
2. **Test webhook manually**:

   ```bash
   curl -X POST -H 'Content-type: application/json' \
     --data '{"text":"Test"}' "$SLACK_WEBHOOK_URL"
   ```

3. **Check app permissions** in Slack workspace

#### Messages Not Appearing

1. **Check channel permissions** - bot must be invited to private channels
2. **Verify webhook is active** in Slack app settings
3. **Check for rate limiting** - Slack has message rate limits

### Email Issues

#### Authentication Failed

1. **Use app-specific passwords** for Gmail with 2FA
2. **Check SMTP settings** for your email provider
3. **Verify credentials**:

   ```bash
   curl --url "smtps://smtp.gmail.com:587" \
     --ssl-reqd --user "$EMAIL_USERNAME:$EMAIL_PASSWORD"
   ```

#### Emails Not Sending

1. **Check SMTP port and encryption**:
   - Port 587: STARTTLS
   - Port 465: SSL/TLS
   - Port 25: Unencrypted (usually blocked)

2. **Verify firewall allows SMTP** traffic

3. **Test with different SMTP servers**:

   ```bash
   # Try different ports
   curl --url "smtps://smtp.gmail.com:465" --ssl-reqd ...
   curl --url "smtp://smtp.gmail.com:587" --starttls ...
   ```

#### Emails Going to Spam

1. **Use proper email headers** (already included in script)
2. **Consider using authenticated SMTP** service (SendGrid, Mailgun)
3. **Add sender to contacts** in recipient's email client

### General Debugging

Enable debug mode for notifications:

```bash
# Add debug output to send_notification function
set -x
send_notification "$provider" "$repo" "$pr_id" "$title" "$author" "merged"
set +x
```

Check notification configuration:

```bash
./test-notifications.sh --config-test
```

## Integration Examples

### CI/CD Pipeline Notifications

Extend for CI/CD integration:

```bash
# Notify when builds complete
if build_successful; then
    send_notification "CI" "$repo" "$build_id" "$commit_message" "$author" "build_passed"
fi
```

### Custom Notification Channels

Add support for other services:

```bash
# Discord webhook
send_discord_notification() {
    local message="$1"
    curl -X POST \
      -H 'Content-Type: application/json' \
      -d "{\"content\":\"$message\"}" \
      "$DISCORD_WEBHOOK_URL"
}

# Microsoft Teams
send_teams_notification() {
    local message="$1"
    curl -X POST \
      -H 'Content-Type: application/json' \
      -d "{\"text\":\"$message\"}" \
      "$TEAMS_WEBHOOK_URL"
}
```
