# test - Test System Functionality and Integrations

The `test` command provides comprehensive testing capabilities for verifying system functionality, integrations, and configurations without affecting production repositories.

## Synopsis

```bash
git-pr-cli test [flags]
```

## Description

The test command offers various testing modes to validate different aspects of the system:

1. **Configuration Testing**: Validates YAML syntax, required fields, and logical consistency
2. **Authentication Testing**: Verifies tokens, permissions, and API connectivity
3. **Integration Testing**: Tests third-party service integrations (Slack, email)
4. **Mock PR Testing**: Simulates PR operations using test repositories or mock data
5. **Performance Testing**: Measures system performance under various loads

## Options

```bash
      --auth                   Test authentication for all providers
      --config                 Test configuration file validation
      --integration            Test external integrations (notifications, webhooks)
  -h, --help                   help for test
      --load-test              Run load testing with multiple concurrent operations
      --mock-prs               Test using mock pull request data
      --notifications          Test notification systems (Slack, email)
      --performance            Run performance benchmarks
      --provider strings       Test specific providers only (github,gitlab,bitbucket)
      --repos strings          Test specific repositories (use with caution)
      --timeout duration       Timeout for individual tests (default 30s)
      --verbose                Show detailed test output
```

## Global Flags

```bash
  -c, --config string   Configuration file path (default "config.yaml")
      --debug           Enable debug logging
      --dry-run         Show what would be tested without executing
```

## Test Categories

### Configuration Tests

```bash
# Test configuration file syntax and structure
git-pr-cli test --config

# Test configuration with verbose output
git-pr-cli test --config --verbose
```

**What it tests:**

- YAML syntax validity
- Required field presence
- Field type validation
- Logical consistency (merge strategies, provider settings)
- Cross-reference validation (repository names, provider configs)

### Authentication Tests

```bash
# Test all provider authentication
git-pr-cli test --auth

# Test specific provider
git-pr-cli test --auth --provider github

# Test with detailed output
git-pr-cli test --auth --verbose --timeout 60s
```

**What it tests:**

- Token validity and format
- API endpoint accessibility
- Rate limit status
- Permission scopes
- Organization/workspace access

### Notification Tests

```bash
# Test all notification systems
git-pr-cli test --notifications

# Test with custom timeout
git-pr-cli test --notifications --timeout 45s
```

**What it tests:**

- Slack webhook connectivity and format
- Email SMTP configuration and delivery
- Message template rendering
- Error handling and fallbacks

### Integration Tests

```bash
# Test external service integrations
git-pr-cli test --integration

# Test with verbose logging
git-pr-cli test --integration --verbose
```

**What it tests:**

- API response parsing
- Error handling mechanisms
- Retry logic functionality
- Timeout behavior

### Performance Tests

```bash
# Run performance benchmarks
git-pr-cli test --performance

# Load testing with concurrent operations
git-pr-cli test --load-test --verbose
```

**What it tests:**

- API response times
- Memory usage patterns
- Concurrent operation handling
- Rate limiting behavior

## Mock Testing

### Mock PR Testing

```bash
# Test using simulated pull request data
git-pr-cli test --mock-prs

# Test with specific scenarios
git-pr-cli test --mock-prs --verbose
```

Mock PR scenarios include:

- Ready-to-merge dependabot PRs
- PRs with failing CI checks
- PRs requiring approvals
- PRs with merge conflicts
- Stale PRs beyond age limits

### Safe Repository Testing

```bash
# Test with actual repositories (use with caution)
git-pr-cli test --repos "test-org/test-repo"

# Always use dry-run for production repositories
git-pr-cli test --repos "prod-org/prod-repo" --dry-run
```

## Example Test Scenarios

### Complete System Test

```bash
# Run all tests
git-pr-cli test --config --auth --notifications --mock-prs --verbose
```

### Pre-deployment Validation

```bash
# Validate before deploying to production
git-pr-cli test --config --auth --integration --performance
```

### Troubleshooting Setup

```bash
# Debug configuration issues
git-pr-cli test --config --verbose --debug

# Debug authentication problems
git-pr-cli test --auth --provider github --verbose --debug
```

### CI/CD Pipeline Testing

```bash
# Automated testing in CI/CD
git-pr-cli test --config --auth --mock-prs --timeout 120s
```

## Test Output Examples

### Configuration Test Output

```text
Configuration Tests
==================
✅ YAML syntax validation passed
✅ Required fields present: pr_filters, repositories, auth
✅ PR filters validation passed
   ├─ allowed_actors: 2 entries
   ├─ skip_labels: 4 entries
   └─ max_age format valid
✅ Repository configuration passed
   ├─ GitHub: 20 repositories
   ├─ GitLab: 3 repositories
   └─ Merge strategies valid
✅ Authentication configuration passed
✅ Notification configuration passed
✅ All configuration tests passed (6/6)
```

### Authentication Test Output

```text
Authentication Tests
===================
✅ GitHub authentication passed
   ├─ Token valid (expires: 2024-07-15)
   ├─ Rate limit: 4,987/5,000 remaining
   ├─ Scopes: repo, read:org
   └─ Organizations: 3 accessible

⚠️  GitLab authentication warning
   ├─ Token valid (expires: never)
   ├─ Rate limit: unlimited
   ├─ Scopes: read_api, write_repository
   └─ Warning: Token has admin privileges (consider using restricted scope)

❌ Bitbucket authentication failed
   ├─ Username valid
   ├─ App password invalid (401 Unauthorized)
   └─ Check BITBUCKET_APP_PASSWORD environment variable

Authentication tests: 2/3 passed, 1 failed
```

### Notification Test Output

```text
Notification Tests
=================
✅ Slack webhook test passed
   ├─ URL accessible
   ├─ Channel #deployments valid
   ├─ Test message sent successfully
   └─ Response time: 234ms

❌ Email SMTP test failed
   ├─ SMTP host accessible
   ├─ Authentication failed (535 Authentication failed)
   └─ Check SMTP_PASSWORD configuration

Notification tests: 1/2 passed, 1 failed
```

### Performance Test Output

```text
Performance Benchmarks
======================
API Response Times (average over 10 requests):
  ├─ GitHub API: 145ms ± 23ms
  ├─ GitLab API: 289ms ± 45ms
  └─ Bitbucket API: 198ms ± 31ms

Memory Usage:
  ├─ Base usage: 12.3 MB
  ├─ Peak usage: 28.7 MB
  └─ Average usage: 18.5 MB

Concurrent Operations (5 parallel):
  ├─ Total time: 2.34s
  ├─ Success rate: 100%
  └─ No rate limit errors

Performance: All benchmarks within acceptable ranges
```

## Mock Data Scenarios

### Dependabot PR Scenario

```json
{
  "number": 123,
  "title": "Bump dependency version from 1.0.0 to 1.0.1",
  "author": "dependabot[bot]",
  "labels": ["dependencies"],
  "status_checks": "passing",
  "approvals": 0,
  "required_approvals": 0,
  "mergeable": true,
  "age_days": 1
}
```

### Failing CI Scenario

```json
{
  "number": 456,
  "title": "Update security dependencies",
  "author": "renovate[bot]",
  "labels": ["security"],
  "status_checks": "failing",
  "mergeable": false,
  "age_days": 2
}
```

### Manual PR Scenario

```json
{
  "number": 789,
  "title": "Feature: Add new API endpoint",
  "author": "developer",
  "labels": ["feature"],
  "status_checks": "passing",
  "approvals": 1,
  "required_approvals": 2,
  "mergeable": true,
  "age_days": 5
}
```

## Advanced Testing

### Load Testing

```bash
# Simulate high repository count
git-pr-cli test --load-test --verbose
```

Load test scenarios:

- 100+ repositories simulation
- Concurrent API operations
- Rate limiting behavior
- Memory usage under load
- Error recovery testing

### Custom Test Configuration

```yaml
# test-config.yaml
test_settings:
  mock_scenarios:
    - ready_prs: 10
    - failing_prs: 3
    - pending_prs: 5

  performance_targets:
    max_response_time: "500ms"
    max_memory_usage: "100MB"
    min_success_rate: 95

  notification_tests:
    slack_timeout: "10s"
    email_timeout: "30s"
```

### Integration with CI/CD

#### GitHub Actions Example

```yaml
name: Test Git PR CLI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Build Git PR CLI
      run: make build

    - name: Run configuration tests
      run: ./git-pr-cli test --config

    - name: Run mock PR tests
      run: ./git-pr-cli test --mock-prs

    - name: Run performance tests
      run: ./git-pr-cli test --performance
      env:
        GITHUB_TOKEN: ${{ secrets.TEST_GITHUB_TOKEN }}
```

#### Jenkins Pipeline Example

```groovy
pipeline {
    agent any

    environment {
        GITHUB_TOKEN = credentials('github-test-token')
    }

    stages {
        stage('Build') {
            steps {
                sh 'make build'
            }
        }

        stage('Test') {
            parallel {
                stage('Configuration') {
                    steps {
                        sh './git-pr-cli test --config --verbose'
                    }
                }

                stage('Authentication') {
                    steps {
                        sh './git-pr-cli test --auth --timeout 60s'
                    }
                }

                stage('Mock PRs') {
                    steps {
                        sh './git-pr-cli test --mock-prs --verbose'
                    }
                }
            }
        }
    }

    post {
        always {
            archiveArtifacts artifacts: 'test-results.xml', allowEmptyArchive: true
        }
    }
}
```

## Exit Codes

- `0`: All tests passed
- `1`: One or more tests failed
- `2`: Configuration error preventing tests
- `3`: Authentication required but not provided
- `4`: Network connectivity issues
- `5`: Test timeout exceeded

## Troubleshooting Test Failures

### Configuration Test Failures

```bash
# Debug YAML syntax errors
git-pr-cli test --config --debug

# Validate with external tools
yq eval . config.yaml
yamllint config.yaml
```

### Authentication Test Failures

```bash
# Test individual providers
git-pr-cli test --auth --provider github --verbose

# Check token permissions manually
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user
```

### Notification Test Failures

```bash
# Test Slack webhook manually
curl -X POST -H 'Content-type: application/json' \
  --data '{"text":"Test message"}' \
  "$SLACK_WEBHOOK_URL"

# Test SMTP connection
telnet smtp.gmail.com 587
```

## Best Practices

1. **Regular Testing**: Include tests in CI/CD pipelines
2. **Incremental Testing**: Test individual components before full system tests
3. **Mock Data**: Use mock PRs for development and testing
4. **Performance Monitoring**: Run performance tests regularly
5. **Security Testing**: Validate token permissions and scopes
6. **Documentation**: Document custom test scenarios and configurations

## Related Commands

- [`validate`](validate.md) - Basic configuration and connectivity validation
- [`check`](check.md) - Live PR status checking (use after successful tests)
- [`setup`](setup.md) - Interactive configuration (test after setup)
- [`stats`](stats.md) - Performance metrics (complement to performance tests)
