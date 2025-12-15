package plugin

import (
	"strings"
	"testing"
)

func TestBuildDashboardURL(t *testing.T) {
	app := &App{
		config: Config{
			GrafanaURL: "http://localhost:3000",
		},
	}

	tests := []struct {
		name     string
		job      Job
		expected string
	}{
		{
			name: "dashboard without variables",
			job: Job{
				DashboardUID: "abc123",
				Slug:         "test-dashboard",
				From:         "now-24h",
				To:           "now",
			},
			expected: "http://localhost:3000/d/abc123/test-dashboard?from=now-24h&to=now",
		},
		{
			name: "dashboard with single variable",
			job: Job{
				DashboardUID: "abc123",
				Slug:         "test-dashboard",
				From:         "now-24h",
				To:           "now",
				Variables: map[string]string{
					"region": "us-east",
				},
			},
			expected: "http://localhost:3000/d/abc123/test-dashboard?from=now-24h&to=now&var-region=us-east",
		},
		{
			name: "dashboard with multiple variables",
			job: Job{
				DashboardUID: "xyz789",
				Slug:         "metrics",
				From:         "now-1h",
				To:           "now",
				Variables: map[string]string{
					"environment": "production",
					"service":     "api",
				},
			},
			expected: "http://localhost:3000/d/xyz789/metrics?from=now-1h&to=now",
		},
		{
			name: "panel with variables",
			job: Job{
				DashboardUID: "panel123",
				Slug:         "panel-dashboard",
				PanelID:      intPtr(5),
				From:         "now-6h",
				To:           "now",
				Variables: map[string]string{
					"host": "server-01",
				},
			},
			expected: "http://localhost:3000/d/panel123/panel-dashboard?viewPanel=5&from=now-6h&to=now&var-host=server-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.buildDashboardURL(app.config.GrafanaURL, tt.job)
			
			// For tests with multiple variables, we can't guarantee order, so check contains
			if len(tt.job.Variables) > 1 {
				// Check that base URL is correct
				baseURL := "http://localhost:3000/d/" + tt.job.DashboardUID + "/" + tt.job.Slug
				if !strings.HasPrefix(result, baseURL) {
					t.Errorf("Expected URL to start with %s, got %s", baseURL, result)
				}
				// Check that all variables are present
				for key, value := range tt.job.Variables {
					expectedVar := "&var-" + key + "=" + value
					if !strings.Contains(result, expectedVar) {
						t.Errorf("Expected URL to contain %s, got %s", expectedVar, result)
					}
				}
			} else {
				// For single or no variables, we can check exact match
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestEmailSenderSendHTML(t *testing.T) {
	// This is a basic test to ensure SendHTML signature is correct
	// Full integration testing would require actual SMTP server
	sender := NewEmailSender("smtp.example.com", "587", "user@example.com", "password", "from@example.com")
	
	if sender == nil {
		t.Fatal("Email sender should not be nil")
	}
	
	// We can't actually send an email in tests without a real SMTP server,
	// but we can verify the function signature accepts the new dashboardURL parameter
	// by attempting to call it (it will fail at SMTP connection, which is expected)
	dashboardURL := "http://localhost:3000/d/abc123/test-dashboard"
	imageData := []byte("fake image data")
	
	// This will fail to send but validates the function signature
	err := sender.SendHTML(
		[]string{"test@example.com"},
		"Test Subject",
		"Test Body",
		imageData,
		"png",
		dashboardURL,
	)
	
	// We expect an error because there's no real SMTP server
	if err == nil {
		t.Log("Note: SendHTML would normally fail without a real SMTP server")
	}
}
