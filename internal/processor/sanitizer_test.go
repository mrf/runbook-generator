package processor

import (
	"testing"

	"github.com/mrf/runbook-generator/internal/history"
)

func TestSanitizer_Passwords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "mysql password short flag",
			input:    "mysql -u root -p'secret123' mydb",
			expected: "mysql -u root -p'<REDACTED>' mydb",
		},
		{
			name:     "mysql password short flag no quotes",
			input:    "mysql -u root -psecret123 mydb",
			expected: "mysql -u root -p<REDACTED> mydb",
		},
		{
			name:     "password long flag",
			input:    "some-command --password=mysecret",
			expected: "some-command --password=<REDACTED>",
		},
		{
			name:     "password long flag with space",
			input:    "some-command --password mysecret",
			expected: "some-command --password <REDACTED>",
		},
		{
			name:     "password long flag quoted",
			input:    `some-command --password="my secret"`,
			expected: `some-command --password="<REDACTED>"`,
		},
	}

	sanitizer := NewSanitizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizer_EnvironmentVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "API key export",
			input:    "export API_KEY=abc123xyz",
			expected: "export API_KEY=<REDACTED>",
		},
		{
			name:     "AWS access key",
			input:    "export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE",
			expected: "export AWS_ACCESS_KEY_ID=<REDACTED>",
		},
		{
			name:     "secret export",
			input:    "export MY_SECRET=supersecretvalue",
			expected: "export MY_SECRET=<REDACTED>",
		},
		{
			name:     "password export",
			input:    "export DB_PASSWORD=mypassword123",
			expected: "export DB_PASSWORD=<REDACTED>",
		},
		{
			name:     "token export",
			input:    "export AUTH_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expected: "export AUTH_TOKEN=<REDACTED>",
		},
	}

	sanitizer := NewSanitizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizer_ConnectionStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "postgres connection string",
			input:    "psql postgres://user:password123@localhost:5432/mydb",
			expected: "psql postgres://user:<REDACTED>@localhost:5432/mydb",
		},
		{
			name:     "mysql connection string",
			input:    "mysql://admin:secret@db.example.com/production",
			expected: "mysql://admin:<REDACTED>@db.example.com/production",
		},
		{
			name:     "redis connection string",
			input:    "redis-cli -u redis://default:mypass@localhost:6379",
			expected: "redis-cli -u redis://default:<REDACTED>@localhost:6379",
		},
	}

	sanitizer := NewSanitizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizer_AuthorizationHeaders(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "bearer token in curl",
			input:    `curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9" https://api.example.com`,
			expected: `curl -H "Authorization: Bearer <REDACTED>" https://api.example.com`,
		},
		{
			name:     "basic auth in curl",
			input:    `curl -H "Authorization: Basic dXNlcjpwYXNz" https://api.example.com`,
			expected: `curl -H "Authorization: Basic <REDACTED>" https://api.example.com`,
		},
	}

	sanitizer := NewSanitizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizer_PrivateKeys(t *testing.T) {
	sanitizer := NewSanitizer()

	input := `echo "-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA0Z3VS5JJcds3xfn/ygWyF8PbnGy...
-----END RSA PRIVATE KEY-----"`

	entry := history.Entry{
		Number:  1,
		Command: input,
	}

	entries, redactions := sanitizer.Process([]history.Entry{entry})

	if len(entries) != 0 {
		t.Errorf("expected private key command to be removed, got %d entries", len(entries))
	}

	if len(redactions) != 1 {
		t.Errorf("expected 1 redaction, got %d", len(redactions))
	}

	if len(redactions) > 0 && redactions[0].PatternName != "private-key" {
		t.Errorf("expected pattern name 'private-key', got %q", redactions[0].PatternName)
	}
}

func TestSanitizer_GitHubTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "github personal access token",
			input:    "gh auth login --with-token ghp_FAKEtestFAKEtestFAKEtestFAKEtestFAKE",
			expected: "gh auth login --with-token <REDACTED>",
		},
		{
			name:     "github pat new format",
			input:    "export GITHUB_TOKEN=github_pat_00FAKETEST_FAKEtestFAKEtestFAKEtestFAKEtestFAKEtestFAKEtestFAKE",
			expected: "export GITHUB_TOKEN=<REDACTED>",
		},
	}

	sanitizer := NewSanitizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizer_DockerSecrets(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "docker env password",
			input:    "docker run -e DB_PASSWORD=secret123 myimage",
			expected: "docker run -e DB_PASSWORD=<REDACTED> myimage",
		},
		{
			name:     "docker env api key",
			input:    "docker run -e API_KEY=abc123 myimage",
			expected: "docker run -e API_KEY=<REDACTED> myimage",
		},
	}

	sanitizer := NewSanitizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizer_PreservesNonSensitive(t *testing.T) {
	inputs := []string{
		"ls -la",
		"git status",
		"docker build -t myimage .",
		"kubectl get pods",
		"npm install express",
		"go build ./...",
		"cat /etc/hosts",
		"echo hello world",
	}

	sanitizer := NewSanitizer()

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			result := sanitizer.SanitizeString(input)
			if result != input {
				t.Errorf("non-sensitive command was modified: got %q, want %q", result, input)
			}
		})
	}
}

// =============================================================
// CRIT-1: JWT Token Tests
// =============================================================
func TestSanitizer_JWTTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "full JWT token in header",
			input:    "curl -H 'X-Custom: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c' https://api.example.com",
			expected: "curl -H 'X-Custom: <REDACTED_JWT>' https://api.example.com",
		},
		{
			name:     "JWT in environment variable",
			input:    "export JWT=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjoiYWRtaW4ifQ.dGVzdHNpZ25hdHVyZQ",
			expected: "export JWT=<REDACTED_JWT>",
		},
		{
			name:     "JWT as command argument",
			input:    "myapp --token eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJteWFwcCJ9.c2lnbmF0dXJl",
			expected: "myapp --token <REDACTED>", // caught by --token flag pattern, still safe
		},
	}

	sanitizer := NewSanitizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// =============================================================
// CRIT-2: Slack, Discord, and Webhook Token Tests
// =============================================================
func TestSanitizer_SlackTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "slack bot token",
			input:    "export SLACK_TOKEN=xoxb-FAKE000000-0000000000000-FakeTestToken000000000000",
			expected: "export SLACK_TOKEN=<REDACTED>",
		},
		{
			name:     "slack user token",
			input:    "curl -H 'Authorization: Bearer xoxp-FAKE000000-FAKE000000-0000000000000-faketesttokenfaketesttoken000000' https://slack.com/api/chat.postMessage",
			expected: "curl -H 'Authorization: Bearer <REDACTED>' https://slack.com/api/chat.postMessage",
		},
		{
			name:     "slack webhook URL",
			input:    "curl -X POST https://hooks.slack.com/services/TFAKETEST/BFAKETEST/FakeWebhookTokenXXXXXX -d '{\"text\":\"Hello\"}'",
			expected: "curl -X POST <REDACTED_SLACK_WEBHOOK> -d '{\"text\":\"Hello\"}'",
		},
	}

	sanitizer := NewSanitizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizer_DiscordWebhooks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "discord webhook URL",
			input:    "curl -X POST https://discord.com/api/webhooks/000000000000000000/FAKEFAKEFAKEtestwebhookFAKEFAKEFAKEFAKEFAKEFAKEFAKE00 -d '{\"content\":\"Hello\"}'",
			expected: "curl -X POST <REDACTED_DISCORD_WEBHOOK> -d '{\"content\":\"Hello\"}'",
		},
		{
			name:     "discordapp webhook URL",
			input:    "wget https://discordapp.com/api/webhooks/000000000000000000/FAKEFAKEFAKEtestwebhookFAKEFAKEFAKEFAKEFAKEFAKEFAKE00",
			expected: "wget <REDACTED_DISCORD_WEBHOOK>",
		},
	}

	sanitizer := NewSanitizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// =============================================================
// CRIT-3: Cloud Provider API Key Tests
// =============================================================
func TestSanitizer_GoogleCloudKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Google API key",
			input:    "export GOOGLE_API_KEY=AIzaFAKE-FAKEtestFAKEtestFAKEtestFAKE00",
			expected: "export GOOGLE_API_KEY=<REDACTED>",
		},
		{
			name:     "Google OAuth token",
			input:    "gcloud auth print-access-token ya29.FAKEtestFAKEtestFAKEtestFAKEtestFAKEtestFAKEtest",
			expected: "gcloud auth print-access-token <REDACTED>",
		},
	}

	sanitizer := NewSanitizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizer_StripeKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Stripe live secret key",
			input:    "export STRIPE_SECRET_KEY=" + "sk_" + "live_0a1b2c3d4e5f6a7b8c9d0e1f2a",
			expected: "export STRIPE_SECRET_KEY=<REDACTED>",
		},
		{
			name:     "Stripe test secret key",
			input:    "curl https://api.stripe.com/v1/charges -u " + "sk_" + "test_0a1b2c3d4e5f6a7b8c9d0e1f2a:",
			expected: "curl https://api.stripe.com/v1/charges -u <REDACTED>:",
		},
	}

	sanitizer := NewSanitizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizer_SendGridMailgun(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "SendGrid API key",
			input:    "export SENDGRID_API_KEY=SG.0000000000000000000000.0000000000000000000000000000000000000000000",
			expected: "export SENDGRID_API_KEY=<REDACTED>",
		},
		{
			name:     "Mailgun API key",
			input:    "curl -s --user 'api:key-00000000000000000000000000000000' https://api.mailgun.net/v3/domains",
			expected: "curl -s --user 'api:<REDACTED>' https://api.mailgun.net/v3/domains",
		},
	}

	sanitizer := NewSanitizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizer_TwilioKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Twilio Account SID",
			input:    "export TWILIO_ACCOUNT_SID=AC00000000000000000000000000000000",
			expected: "export TWILIO_ACCOUNT_SID=<REDACTED>",
		},
		{
			name:     "Twilio API key",
			input:    "export TWILIO_API_KEY=SK00000000000000000000000000000000",
			expected: "export TWILIO_API_KEY=<REDACTED>",
		},
	}

	sanitizer := NewSanitizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizer_DigitalOceanTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "DigitalOcean API token",
			input:    "export DO_TOKEN=dop_v1_0000000000000000000000000000000000000000000000000000000000000000",
			expected: "export DO_TOKEN=<REDACTED>",
		},
	}

	sanitizer := NewSanitizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizer_NPMPyPITokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "NPM token",
			input:    "npm config set //registry.npmjs.org/:_authToken npm_000000000000000000000000000000000000",
			expected: "npm config set //registry.npmjs.org/:_authToken <REDACTED>",
		},
		{
			name:     "PyPI token",
			input:    "export TWINE_PASSWORD=pypi-00000000000000000000000000000000000000000000000000",
			expected: "export TWINE_PASSWORD=<REDACTED>",
		},
	}

	sanitizer := NewSanitizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}
