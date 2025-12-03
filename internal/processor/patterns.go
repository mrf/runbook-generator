package processor

import "regexp"

// Pattern defines a regex pattern for detecting and redacting sensitive data.
type Pattern struct {
	Name        string
	Regex       *regexp.Regexp
	Replacement string
	FullRemove  bool // If true, remove entire command instead of redacting
}

// DefaultPatterns returns the built-in secret detection patterns.
// NOTE: Order matters! More specific patterns should come before general ones.
func DefaultPatterns() []Pattern {
	return []Pattern{
		// Password flags - long flags first (more specific)
		// Handle quoted values (may contain spaces)
		{
			Name:        "password-flag-quoted-double",
			Regex:       regexp.MustCompile(`(--password[=\s]+)"([^"]+)"`),
			Replacement: `${1}"<REDACTED>"`,
		},
		{
			Name:        "password-flag-quoted-single",
			Regex:       regexp.MustCompile(`(--password[=\s]+)'([^']+)'`),
			Replacement: `${1}'<REDACTED>'`,
		},
		// Handle unquoted values
		{
			Name:        "password-flag-unquoted",
			Regex:       regexp.MustCompile(`(--password[=\s]+)([^'"'\s]+)`),
			Replacement: "${1}<REDACTED>",
		},
		{
			Name:        "passwd-flag",
			Regex:       regexp.MustCompile(`(--passwd[=\s]+)(['"]?)([^'"'\s]+)(['"]?)`),
			Replacement: "${1}${2}<REDACTED>${4}",
		},
		// MySQL short -p flag (must NOT be preceded by another dash)
		{
			Name:        "mysql-password",
			Regex:       regexp.MustCompile(`(\s-p)(['"]?)([^'"'\s-][^'"'\s]*)(['"]?)`),
			Replacement: "${1}${2}<REDACTED>${4}",
		},

		// Token and API key flags
		{
			Name:        "token-flag",
			Regex:       regexp.MustCompile(`(--token[=\s]+['"]?)([^'"'\s]+)(['"]?)`),
			Replacement: "${1}<REDACTED>${3}",
		},
		{
			Name:        "api-key-flag",
			Regex:       regexp.MustCompile(`(--api-key[=\s]+['"]?)([^'"'\s]+)(['"]?)`),
			Replacement: "${1}<REDACTED>${3}",
		},
		{
			Name:        "secret-flag",
			Regex:       regexp.MustCompile(`(--secret[=\s]+['"]?)([^'"'\s]+)(['"]?)`),
			Replacement: "${1}<REDACTED>${3}",
		},

		// Environment variable exports
		{
			Name:        "api-key-export",
			Regex:       regexp.MustCompile(`(export\s+[A-Z_]*API[_]?KEY\s*=\s*['"]?)([^'"'\s]+)(['"]?)`),
			Replacement: "${1}<REDACTED>${3}",
		},
		{
			Name:        "secret-export",
			Regex:       regexp.MustCompile(`(export\s+[A-Z_]*SECRET[A-Z_]*\s*=\s*['"]?)([^'"'\s]+)(['"]?)`),
			Replacement: "${1}<REDACTED>${3}",
		},
		{
			Name:        "password-export",
			Regex:       regexp.MustCompile(`(export\s+[A-Z_]*PASS(?:WORD)?[A-Z_]*\s*=\s*['"]?)([^'"'\s]+)(['"]?)`),
			Replacement: "${1}<REDACTED>${3}",
		},
		{
			Name:        "token-export",
			Regex:       regexp.MustCompile(`(export\s+[A-Z_]*TOKEN\s*=\s*['"]?)([^'"'\s]+)(['"]?)`),
			Replacement: "${1}<REDACTED>${3}",
		},
		{
			Name:        "credentials-export",
			Regex:       regexp.MustCompile(`(export\s+[A-Z_]*CRED(?:ENTIAL)?S?[A-Z_]*\s*=\s*['"]?)([^'"'\s]+)(['"]?)`),
			Replacement: "${1}<REDACTED>${3}",
		},

		// AWS credentials
		{
			Name:        "aws-access-key-id",
			Regex:       regexp.MustCompile(`(AWS_ACCESS_KEY_ID\s*=\s*['"]?)([A-Z0-9]{20})(['"]?)`),
			Replacement: "${1}<REDACTED>${3}",
		},
		{
			Name:        "aws-secret-key",
			Regex:       regexp.MustCompile(`(AWS_SECRET_ACCESS_KEY\s*=\s*['"]?)([A-Za-z0-9/+=]{40})(['"]?)`),
			Replacement: "${1}<REDACTED>${3}",
		},
		{
			Name:        "aws-session-token",
			Regex:       regexp.MustCompile(`(AWS_SESSION_TOKEN\s*=\s*['"]?)([^'"'\s]+)(['"]?)`),
			Replacement: "${1}<REDACTED>${3}",
		},

		// Connection strings
		{
			Name:        "connection-string-password",
			Regex:       regexp.MustCompile(`(://[^:]+:)([^@]+)(@)`),
			Replacement: "${1}<REDACTED>${3}",
		},

		// Authorization headers
		{
			Name:        "bearer-token",
			Regex:       regexp.MustCompile(`(Authorization:\s*Bearer\s+)([^\s'"]+)`),
			Replacement: "${1}<REDACTED>",
		},
		{
			Name:        "basic-auth",
			Regex:       regexp.MustCompile(`(Authorization:\s*Basic\s+)([^\s'"]+)`),
			Replacement: "${1}<REDACTED>",
		},
		{
			Name:        "auth-header-h-flag",
			Regex:       regexp.MustCompile(`(-H\s+['"]?Authorization:\s*(?:Bearer|Basic)\s+)([^'"'\s]+)(['"]?)`),
			Replacement: "${1}<REDACTED>${3}",
		},

		// Private keys - full removal (HIGH-2: expanded to all key types)
		// Matches: RSA, EC, OPENSSH, ENCRYPTED, DSA, and generic PRIVATE KEY
		{
			Name:        "private-key",
			Regex:       regexp.MustCompile(`-----BEGIN\s+(?:RSA\s+|EC\s+|OPENSSH\s+|ENCRYPTED\s+|DSA\s+)?PRIVATE\s+KEY-----`),
			Replacement: "",
			FullRemove:  true,
		},
		// PGP private keys
		{
			Name:        "pgp-private-key",
			Regex:       regexp.MustCompile(`-----BEGIN\s+PGP\s+PRIVATE\s+KEY\s+BLOCK-----`),
			Replacement: "",
			FullRemove:  true,
		},

		// GitHub tokens
		{
			Name:        "github-token",
			Regex:       regexp.MustCompile(`(gh[ps]_[A-Za-z0-9]{36,})`),
			Replacement: "<REDACTED>",
		},
		{
			Name:        "github-pat",
			Regex:       regexp.MustCompile(`(github_pat_[A-Za-z0-9_]{22,})`),
			Replacement: "<REDACTED>",
		},

		// =============================================================
		// CRIT-1: JWT Tokens (full format with header.payload.signature)
		// =============================================================
		{
			Name:        "jwt-token",
			Regex:       regexp.MustCompile(`\beyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b`),
			Replacement: "<REDACTED_JWT>",
		},

		// =============================================================
		// CRIT-2: Slack, Discord, and Webhook Tokens
		// =============================================================
		// Slack Bot tokens (xoxb-)
		{
			Name:        "slack-bot-token",
			Regex:       regexp.MustCompile(`xoxb-[0-9]+-[0-9]+-[A-Za-z0-9]+`),
			Replacement: "<REDACTED>",
		},
		// Slack User tokens (xoxp-)
		{
			Name:        "slack-user-token",
			Regex:       regexp.MustCompile(`xoxp-[0-9]+-[0-9]+-[0-9]+-[A-Za-z0-9]+`),
			Replacement: "<REDACTED>",
		},
		// Slack App tokens (xapp-)
		{
			Name:        "slack-app-token",
			Regex:       regexp.MustCompile(`xapp-[0-9]+-[A-Za-z0-9]+-[0-9]+-[A-Za-z0-9]+`),
			Replacement: "<REDACTED>",
		},
		// Slack Refresh tokens (xoxr-)
		{
			Name:        "slack-refresh-token",
			Regex:       regexp.MustCompile(`xoxr-[0-9]+-[A-Za-z0-9]+`),
			Replacement: "<REDACTED>",
		},
		// Slack Webhook URLs
		{
			Name:        "slack-webhook",
			Regex:       regexp.MustCompile(`https://hooks\.slack\.com/services/T[A-Z0-9]+/B[A-Z0-9]+/[A-Za-z0-9]+`),
			Replacement: "<REDACTED_SLACK_WEBHOOK>",
		},
		// Discord Webhook URLs
		{
			Name:        "discord-webhook",
			Regex:       regexp.MustCompile(`https://discord(?:app)?\.com/api/webhooks/[0-9]+/[A-Za-z0-9_-]+`),
			Replacement: "<REDACTED_DISCORD_WEBHOOK>",
		},
		// Generic webhook URLs with tokens/secrets in path
		{
			Name:        "generic-webhook-secret",
			Regex:       regexp.MustCompile(`(https?://[^/]+/webhooks?/)[A-Za-z0-9_-]{20,}`),
			Replacement: "${1}<REDACTED>",
		},

		// =============================================================
		// CRIT-3: Cloud Provider API Keys and Tokens
		// =============================================================
		// Google Cloud API keys (AIza...)
		{
			Name:        "gcp-api-key",
			Regex:       regexp.MustCompile(`AIza[A-Za-z0-9_-]{35}`),
			Replacement: "<REDACTED>",
		},
		// Google OAuth tokens
		{
			Name:        "google-oauth",
			Regex:       regexp.MustCompile(`ya29\.[A-Za-z0-9_-]+`),
			Replacement: "<REDACTED>",
		},
		// Azure Storage Account keys
		{
			Name:        "azure-storage-key",
			Regex:       regexp.MustCompile(`(?i)(AccountKey\s*=\s*)([A-Za-z0-9+/=]{88})`),
			Replacement: "${1}<REDACTED>",
		},
		// Azure Connection Strings
		{
			Name:        "azure-connection-string",
			Regex:       regexp.MustCompile(`(?i)(DefaultEndpointsProtocol=https?;AccountName=[^;]+;AccountKey=)([A-Za-z0-9+/=]+)`),
			Replacement: "${1}<REDACTED>",
		},
		// Azure SAS tokens
		{
			Name:        "azure-sas-token",
			Regex:       regexp.MustCompile(`(\?|&)(sig|sv|ss|srt|sp|se|st|spr|sr)=[^&\s]+`),
			Replacement: "${1}${2}=<REDACTED>",
		},
		// DigitalOcean tokens
		{
			Name:        "digitalocean-token",
			Regex:       regexp.MustCompile(`dop_v1_[a-f0-9]{64}`),
			Replacement: "<REDACTED>",
		},
		// DigitalOcean OAuth tokens
		{
			Name:        "digitalocean-oauth",
			Regex:       regexp.MustCompile(`doo_v1_[a-f0-9]{64}`),
			Replacement: "<REDACTED>",
		},
		// Stripe API keys (secret and publishable)
		{
			Name:        "stripe-secret-key",
			Regex:       regexp.MustCompile(`sk_live_[A-Za-z0-9]{24,}`),
			Replacement: "<REDACTED>",
		},
		{
			Name:        "stripe-test-key",
			Regex:       regexp.MustCompile(`sk_test_[A-Za-z0-9]{24,}`),
			Replacement: "<REDACTED>",
		},
		{
			Name:        "stripe-restricted-key",
			Regex:       regexp.MustCompile(`rk_live_[A-Za-z0-9]{24,}`),
			Replacement: "<REDACTED>",
		},
		// Twilio Account SID and Auth Token
		{
			Name:        "twilio-account-sid",
			Regex:       regexp.MustCompile(`AC[a-f0-9]{32}`),
			Replacement: "<REDACTED>",
		},
		{
			Name:        "twilio-api-key",
			Regex:       regexp.MustCompile(`SK[a-f0-9]{32}`),
			Replacement: "<REDACTED>",
		},
		// SendGrid API key
		{
			Name:        "sendgrid-api-key",
			Regex:       regexp.MustCompile(`SG\.[A-Za-z0-9_-]{22}\.[A-Za-z0-9_-]{43}`),
			Replacement: "<REDACTED>",
		},
		// Mailgun API key
		{
			Name:        "mailgun-api-key",
			Regex:       regexp.MustCompile(`key-[a-f0-9]{32}`),
			Replacement: "<REDACTED>",
		},
		// NPM tokens
		{
			Name:        "npm-token",
			Regex:       regexp.MustCompile(`npm_[A-Za-z0-9]{36,}`),
			Replacement: "<REDACTED>",
		},
		// PyPI tokens
		{
			Name:        "pypi-token",
			Regex:       regexp.MustCompile(`pypi-[A-Za-z0-9_-]{50,}`),
			Replacement: "<REDACTED>",
		},
		// Heroku API key
		{
			Name:        "heroku-api-key",
			Regex:       regexp.MustCompile(`(?i)(HEROKU_API_KEY\s*=\s*['"]?)([a-f0-9-]{36})(['"]?)`),
			Replacement: "${1}<REDACTED>${3}",
		},
		// Shopify tokens
		{
			Name:        "shopify-access-token",
			Regex:       regexp.MustCompile(`shpat_[a-f0-9]{32}`),
			Replacement: "<REDACTED>",
		},
		{
			Name:        "shopify-shared-secret",
			Regex:       regexp.MustCompile(`shpss_[a-f0-9]{32}`),
			Replacement: "<REDACTED>",
		},
		// Square tokens
		{
			Name:        "square-access-token",
			Regex:       regexp.MustCompile(`sq0atp-[A-Za-z0-9_-]{22}`),
			Replacement: "<REDACTED>",
		},
		{
			Name:        "square-oauth-secret",
			Regex:       regexp.MustCompile(`sq0csp-[A-Za-z0-9_-]{43}`),
			Replacement: "<REDACTED>",
		},
		// Datadog API key
		{
			Name:        "datadog-api-key",
			Regex:       regexp.MustCompile(`(?i)(DD_API_KEY\s*=\s*['"]?)([a-f0-9]{32})(['"]?)`),
			Replacement: "${1}<REDACTED>${3}",
		},
		// New Relic API key
		{
			Name:        "newrelic-api-key",
			Regex:       regexp.MustCompile(`NRAK-[A-Z0-9]{27}`),
			Replacement: "<REDACTED>",
		},
		// Vault tokens (HashiCorp)
		{
			Name:        "vault-token",
			Regex:       regexp.MustCompile(`(?i)(VAULT_TOKEN\s*=\s*['"]?)([shrs]\.[A-Za-z0-9_-]+)(['"]?)`),
			Replacement: "${1}<REDACTED>${3}",
		},
		// MongoDB connection strings
		{
			Name:        "mongodb-connection-string",
			Regex:       regexp.MustCompile(`mongodb(?:\+srv)?://[^:]+:([^@]+)@`),
			Replacement: "mongodb://[user]:<REDACTED>@",
		},

		// Generic secrets in common formats
		// NOTE: Limited \w{0,20} to prevent ReDoS (HIGH-1 fix)
		{
			Name:        "generic-secret-assignment",
			Regex:       regexp.MustCompile(`(?i)((?:secret|password|passwd|pwd|token|api_key|apikey|auth)[_-]?\w{0,20}\s{0,3}[=:]\s{0,3}['"]?)([^'"'\s]{8,})(['"]?)`),
			Replacement: "${1}<REDACTED>${3}",
		},

		// Inline variable assignments before commands
		{
			Name:        "inline-secret-var",
			Regex:       regexp.MustCompile(`(\b(?:PASSWORD|SECRET|TOKEN|API_KEY)\s*=\s*['"]?)([^'"'\s]+)(['"]?\s)`),
			Replacement: "${1}<REDACTED>${3}",
		},

		// curl with data containing passwords
		{
			Name:        "curl-password-data",
			Regex:       regexp.MustCompile(`(-d\s+['"]?[^'"]*(?:password|passwd|secret|token)['"]*\s*[=:]\s*['"]?)([^'"&\s]+)(['"]?)`),
			Replacement: "${1}<REDACTED>${3}",
		},

		// Docker/container secrets
		{
			Name:        "docker-secret-env",
			Regex:       regexp.MustCompile(`(-e\s+[A-Z_]*(?:PASSWORD|SECRET|TOKEN|API_KEY)[A-Z_]*=)([^\s]+)`),
			Replacement: "${1}<REDACTED>",
		},

		// kubectl secret data
		{
			Name:        "kubectl-secret",
			Regex:       regexp.MustCompile(`(--from-literal=[A-Za-z_-]*(?:password|secret|token|key)[A-Za-z_-]*=)([^\s]+)`),
			Replacement: "${1}<REDACTED>",
		},
	}
}
