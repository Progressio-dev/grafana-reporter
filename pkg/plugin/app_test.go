package plugin

import (
	"testing"
)

func TestJobValidation(t *testing.T) {
	tests := []struct {
		name    string
		job     Job
		wantErr bool
	}{
		{
			name: "valid job",
			job: Job{
				ID:           "test-1",
				Cron:         "0 9 * * *",
				DashboardUID: "abc123",
				Slug:         "test-dashboard",
				From:         "now-24h",
				To:           "now",
				Width:        1920,
				Height:       1080,
				Scale:        1,
				Format:       "png",
				Recipients:   []string{"test@example.com"},
				Subject:      "Test Report",
				Body:         "Test body",
			},
			wantErr: false,
		},
		{
			name: "job with panel ID",
			job: Job{
				ID:           "test-2",
				Cron:         "0 * * * *",
				DashboardUID: "xyz789",
				Slug:         "metrics",
				PanelID:      intPtr(2),
				From:         "now-1h",
				To:           "now",
				Width:        800,
				Height:       600,
				Scale:        2,
				Format:       "pdf",
				Recipients:   []string{"admin@example.com", "user@example.com"},
				Subject:      "Hourly Metrics",
				Body:         "Hourly metrics report",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - ensure required fields are set
			if tt.job.ID == "" {
				t.Error("Job ID should not be empty")
			}
			if tt.job.Cron == "" {
				t.Error("Cron expression should not be empty")
			}
			if tt.job.DashboardUID == "" {
				t.Error("Dashboard UID should not be empty")
			}
			if len(tt.job.Recipients) == 0 {
				t.Error("Recipients should not be empty")
			}
		})
	}
}

func TestEmailSenderCreation(t *testing.T) {
	sender := NewEmailSender("smtp.example.com", "587", "user@example.com", "password", "from@example.com")
	
	if sender == nil {
		t.Error("Email sender should not be nil")
	}
	
	if sender.host != "smtp.example.com" {
		t.Errorf("Expected host to be smtp.example.com, got %s", sender.host)
	}
	
	if sender.port != "587" {
		t.Errorf("Expected port to be 587, got %s", sender.port)
	}
}

func intPtr(i int) *int {
	return &i
}
