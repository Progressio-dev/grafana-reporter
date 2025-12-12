package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
	"github.com/robfig/cron/v3"
)

// Job represents a scheduled report job
type Job struct {
	ID           string   `json:"id"`
	Cron         string   `json:"cron"`
	DashboardUID string   `json:"dashboardUid"`
	Slug         string   `json:"slug"`
	PanelID      *int     `json:"panelId,omitempty"`
	From         string   `json:"from"`
	To           string   `json:"to"`
	Width        int      `json:"width"`
	Height       int      `json:"height"`
	Scale        int      `json:"scale"`
	Format       string   `json:"format"` // png or pdf
	Recipients   []string `json:"recipients"`
	Subject      string   `json:"subject"`
	Body         string   `json:"body"`
}

// App implements the backend plugin
type App struct {
	backend.CallResourceHandler
	
	scheduler *cron.Cron
	jobs      map[string]Job
	cronIDs   map[string]cron.EntryID
	jobsFile  string
	mu        sync.RWMutex
	
	grafanaURL string
	apiKey     string
}

// NewApp creates a new App instance
func NewApp(ctx context.Context, settings backend.AppInstanceSettings) (instancemgmt.Instance, error) {
	log.DefaultLogger.Info("Creating new app instance")
	
	app := &App{
		scheduler: cron.New(),
		jobs:      make(map[string]Job),
		cronIDs:   make(map[string]cron.EntryID),
		jobsFile:  "data/jobs.json",
	}
	
	// Parse settings
	if settings.JSONData != nil {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(settings.JSONData, &jsonData); err == nil {
			if url, ok := jsonData["grafanaUrl"].(string); ok {
				app.grafanaURL = url
			}
		}
	}
	
	// Get API key from secure JSON data
	if settings.DecryptedSecureJSONData != nil {
		if key, ok := settings.DecryptedSecureJSONData["apiKey"]; ok {
			app.apiKey = key
		}
	}
	
	// Set default Grafana URL if not configured
	if app.grafanaURL == "" {
		app.grafanaURL = os.Getenv("GRAFANA_URL")
		if app.grafanaURL == "" {
			app.grafanaURL = "http://localhost:3000"
		}
	}
	
	// Load jobs from file
	if err := app.loadJobs(); err != nil {
		log.DefaultLogger.Warn("Failed to load jobs", "error", err)
	}
	
	// Start scheduler
	app.scheduler.Start()
	
	// Schedule all loaded jobs
	for _, job := range app.jobs {
		if err := app.scheduleJob(job); err != nil {
			log.DefaultLogger.Error("Failed to schedule job", "id", job.ID, "error", err)
		}
	}
	
	// Set up resource handler
	mux := http.NewServeMux()
	mux.HandleFunc("/jobs", app.handleJobs)
	mux.HandleFunc("/jobs/", app.handleJobByID)
	mux.HandleFunc("/test-email", app.handleTestEmail)
	app.CallResourceHandler = httpadapter.New(mux)
	
	return app, nil
}

// Dispose cleans up resources
func (app *App) Dispose() {
	log.DefaultLogger.Info("Disposing app instance")
	if app.scheduler != nil {
		app.scheduler.Stop()
	}
}

// loadJobs loads jobs from the JSON file
func (app *App) loadJobs() error {
	app.mu.Lock()
	defer app.mu.Unlock()
	
	// Create data directory if it doesn't exist
	dir := filepath.Dir(app.jobsFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	
	// Check if file exists
	if _, err := os.Stat(app.jobsFile); os.IsNotExist(err) {
		// Create empty jobs file
		if err := os.WriteFile(app.jobsFile, []byte("[]"), 0644); err != nil {
			return fmt.Errorf("failed to create jobs file: %w", err)
		}
		return nil
	}
	
	data, err := os.ReadFile(app.jobsFile)
	if err != nil {
		return fmt.Errorf("failed to read jobs file: %w", err)
	}
	
	var jobs []Job
	if err := json.Unmarshal(data, &jobs); err != nil {
		return fmt.Errorf("failed to parse jobs file: %w", err)
	}
	
	for _, job := range jobs {
		app.jobs[job.ID] = job
	}
	
	log.DefaultLogger.Info("Loaded jobs", "count", len(app.jobs))
	return nil
}

// saveJobs saves jobs to the JSON file
func (app *App) saveJobs() error {
	app.mu.RLock()
	jobs := make([]Job, 0, len(app.jobs))
	for _, job := range app.jobs {
		jobs = append(jobs, job)
	}
	app.mu.RUnlock()
	
	data, err := json.MarshalIndent(jobs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal jobs: %w", err)
	}
	
	if err := os.WriteFile(app.jobsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write jobs file: %w", err)
	}
	
	return nil
}

// scheduleJob schedules a job with the cron scheduler
func (app *App) scheduleJob(job Job) error {
	app.mu.Lock()
	defer app.mu.Unlock()
	
	// Remove existing schedule if any
	if entryID, ok := app.cronIDs[job.ID]; ok {
		app.scheduler.Remove(entryID)
		delete(app.cronIDs, job.ID)
	}
	
	// Add new schedule
	entryID, err := app.scheduler.AddFunc(job.Cron, func() {
		if err := app.executeJob(job); err != nil {
			log.DefaultLogger.Error("Failed to execute job", "id", job.ID, "error", err)
		}
	})
	
	if err != nil {
		return fmt.Errorf("failed to schedule job: %w", err)
	}
	
	app.cronIDs[job.ID] = entryID
	log.DefaultLogger.Info("Scheduled job", "id", job.ID, "cron", job.Cron)
	
	return nil
}

// unscheduleJob removes a job from the scheduler
func (app *App) unscheduleJob(jobID string) {
	app.mu.Lock()
	defer app.mu.Unlock()
	
	if entryID, ok := app.cronIDs[jobID]; ok {
		app.scheduler.Remove(entryID)
		delete(app.cronIDs, jobID)
		log.DefaultLogger.Info("Unscheduled job", "id", jobID)
	}
}

// executeJob executes a scheduled job
func (app *App) executeJob(job Job) error {
	log.DefaultLogger.Info("Executing job", "id", job.ID)
	
	// Render the report
	imageData, err := app.renderReport(job)
	if err != nil {
		return fmt.Errorf("failed to render report: %w", err)
	}
	
	// Send email
	if err := app.sendEmail(job, imageData); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	
	log.DefaultLogger.Info("Job executed successfully", "id", job.ID)
	return nil
}

// renderReport renders a dashboard or panel to PNG/PDF
func (app *App) renderReport(job Job) ([]byte, error) {
	// Build render URL
	var renderURL string
	if job.PanelID != nil {
		// Render single panel
		renderURL = fmt.Sprintf("%s/render/d-solo/%s/%s?panelId=%d&from=%s&to=%s&width=%d&height=%d&scale=%d&tz=UTC",
			app.grafanaURL, job.DashboardUID, job.Slug, *job.PanelID, job.From, job.To, job.Width, job.Height, job.Scale)
	} else {
		// Render full dashboard
		renderURL = fmt.Sprintf("%s/render/d/%s/%s?from=%s&to=%s&width=%d&height=%d&scale=%d&kiosk&tz=UTC",
			app.grafanaURL, job.DashboardUID, job.Slug, job.From, job.To, job.Width, job.Height, job.Scale)
	}
	
	log.DefaultLogger.Debug("Rendering report", "url", renderURL, "format", job.Format)
	
	// Create HTTP request
	req, err := http.NewRequest("GET", renderURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add authorization header
	if app.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+app.apiKey)
	}
	
	// Execute request
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("render request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	return data, nil
}

// sendEmail sends an email with the rendered report
func (app *App) sendEmail(job Job, attachment []byte) error {
	log.DefaultLogger.Info("Sending email", "recipients", job.Recipients, "subject", job.Subject)
	
	// Get SMTP configuration from environment
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := os.Getenv("SMTP_FROM")
	
	if smtpHost == "" {
		return fmt.Errorf("SMTP_HOST not configured")
	}
	
	if smtpFrom == "" {
		smtpFrom = smtpUser
	}
	
	if smtpPort == "" {
		smtpPort = "587"
	}
	
	// Create email sender
	sender := NewEmailSender(smtpHost, smtpPort, smtpUser, smtpPass, smtpFrom)
	
	// Determine attachment filename
	filename := fmt.Sprintf("report-%s.%s", time.Now().Format("2006-01-02-150405"), job.Format)
	
	// Send email
	return sender.Send(job.Recipients, job.Subject, job.Body, attachment, filename)
}

// HTTP handlers

func (app *App) handleJobs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.listJobs(w, r)
	case http.MethodPost:
		app.createJob(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) handleJobByID(w http.ResponseWriter, r *http.Request) {
	// Extract job ID from path
	jobID := r.URL.Path[len("/jobs/"):]
	
	switch r.Method {
	case http.MethodGet:
		app.getJob(w, r, jobID)
	case http.MethodPut:
		app.updateJob(w, r, jobID)
	case http.MethodDelete:
		app.deleteJob(w, r, jobID)
	case http.MethodPost:
		// Support /jobs/{id}/execute endpoint
		if r.URL.Path == fmt.Sprintf("/jobs/%s/execute", jobID) {
			app.executeJobHandler(w, r, jobID)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) listJobs(w http.ResponseWriter, r *http.Request) {
	app.mu.RLock()
	jobs := make([]Job, 0, len(app.jobs))
	for _, job := range app.jobs {
		jobs = append(jobs, job)
	}
	app.mu.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func (app *App) getJob(w http.ResponseWriter, r *http.Request, jobID string) {
	app.mu.RLock()
	job, ok := app.jobs[jobID]
	app.mu.RUnlock()
	
	if !ok {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func (app *App) createJob(w http.ResponseWriter, r *http.Request) {
	var job Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Generate ID if not provided
	if job.ID == "" {
		job.ID = fmt.Sprintf("job-%d", time.Now().UnixNano())
	}
	
	// Validate cron expression
	if _, err := cron.ParseStandard(job.Cron); err != nil {
		http.Error(w, fmt.Sprintf("Invalid cron expression: %v", err), http.StatusBadRequest)
		return
	}
	
	// Add job
	app.mu.Lock()
	app.jobs[job.ID] = job
	app.mu.Unlock()
	
	// Schedule job
	if err := app.scheduleJob(job); err != nil {
		http.Error(w, fmt.Sprintf("Failed to schedule job: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Save to file
	if err := app.saveJobs(); err != nil {
		log.DefaultLogger.Error("Failed to save jobs", "error", err)
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)
}

func (app *App) updateJob(w http.ResponseWriter, r *http.Request, jobID string) {
	var job Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Ensure ID matches
	job.ID = jobID
	
	// Validate cron expression
	if _, err := cron.ParseStandard(job.Cron); err != nil {
		http.Error(w, fmt.Sprintf("Invalid cron expression: %v", err), http.StatusBadRequest)
		return
	}
	
	// Check if job exists
	app.mu.RLock()
	_, exists := app.jobs[jobID]
	app.mu.RUnlock()
	
	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}
	
	// Update job
	app.mu.Lock()
	app.jobs[jobID] = job
	app.mu.Unlock()
	
	// Reschedule job
	if err := app.scheduleJob(job); err != nil {
		http.Error(w, fmt.Sprintf("Failed to schedule job: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Save to file
	if err := app.saveJobs(); err != nil {
		log.DefaultLogger.Error("Failed to save jobs", "error", err)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func (app *App) deleteJob(w http.ResponseWriter, r *http.Request, jobID string) {
	// Check if job exists
	app.mu.RLock()
	_, exists := app.jobs[jobID]
	app.mu.RUnlock()
	
	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}
	
	// Unschedule job
	app.unscheduleJob(jobID)
	
	// Remove job
	app.mu.Lock()
	delete(app.jobs, jobID)
	app.mu.Unlock()
	
	// Save to file
	if err := app.saveJobs(); err != nil {
		log.DefaultLogger.Error("Failed to save jobs", "error", err)
	}
	
	w.WriteHeader(http.StatusNoContent)
}

func (app *App) executeJobHandler(w http.ResponseWriter, r *http.Request, jobID string) {
	// Get job
	app.mu.RLock()
	job, exists := app.jobs[jobID]
	app.mu.RUnlock()
	
	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}
	
	// Execute job asynchronously
	go func() {
		if err := app.executeJob(job); err != nil {
			log.DefaultLogger.Error("Failed to execute job", "id", jobID, "error", err)
		}
	}()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Job execution started",
	})
}

func (app *App) handleTestEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		Recipients []string `json:"recipients"`
		Subject    string   `json:"subject"`
		Body       string   `json:"body"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Create a test message
	testMessage := []byte("This is a test email from Grafana Reporter plugin.")
	
	testJob := Job{
		Recipients: req.Recipients,
		Subject:    req.Subject,
		Body:       req.Body,
		Format:     "test", // Using "test" format to distinguish from actual report formats
	}
	
	if err := app.sendEmail(testJob, testMessage); err != nil {
		http.Error(w, fmt.Sprintf("Failed to send test email: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Test email sent successfully",
	})
}

// CheckHealth handles health check requests
func (app *App) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Info("Checking health")
	
	status := backend.HealthStatusOk
	message := "Plugin is healthy"
	
	// Check if scheduler is running
	if app.scheduler == nil {
		status = backend.HealthStatusError
		message = "Scheduler is not running"
	}
	
	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}
