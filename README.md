# Grafana Reporter Plugin

Schedule PDF/PNG report generation and email sending using Grafana OSS and grafana-image-renderer.

## Features

- 📅 **Cron-based Scheduler**: Schedule reports using cron expressions
- 📊 **Dashboard & Panel Rendering**: Render full dashboards or individual panels
- 📧 **Email Delivery**: Automatically send rendered reports via email
- 🖼️ **Multiple Formats**: Support for PNG and PDF output
- 🔒 **Secure**: Uses Grafana's authorization system
- 🎯 **Dashboard Selector**: Select dashboards from dropdown with folder hierarchy (no need to manually enter UID/slug)
- 🔧 **Variable Support**: Set dashboard variables for each report (key=value format)

## Requirements

- Grafana 9.0+
- grafana-image-renderer plugin installed and configured
- SMTP server for email sending

## Installation

1. Copy the plugin to your Grafana plugins directory:
   ```bash
   cp -r dist /var/lib/grafana/plugins/progressio-grafanareporter-app
   ```

2. Configure SMTP settings via environment variables:
   ```bash
   export SMTP_HOST=smtp.gmail.com
   export SMTP_PORT=587
   export SMTP_USER=your-email@gmail.com
   export SMTP_PASS=your-password
   export SMTP_FROM=your-email@gmail.com
   ```

3. Configure Grafana API key:
   - Create a service account in Grafana with appropriate permissions
   - Generate an API token
   - Configure the plugin with the API token in Grafana settings

4. Restart Grafana:
   ```bash
   systemctl restart grafana-server
   ```

## Building

### Backend

```bash
# Install dependencies
go mod download

# Build the backend binary
go build -o dist/gpx_grafana-reporter ./pkg
```

### Frontend

```bash
# Install dependencies
npm install

# Build for production
npm run build

# Or for development with live reload
npm run dev
```

## Usage

1. Navigate to the Grafana Reporter plugin page in Grafana
2. Click "Create Job" to add a new scheduled report
3. Configure:
   - Cron expression (e.g., `0 9 * * *` for daily at 9 AM)
   - Dashboard: Select from dropdown (shows all dashboards organized by folder)
   - Optional: Panel ID for single panel rendering
   - Time range (from/to)
   - Optional: Variables (key=value format, one per line)
   - Rendering dimensions (width, height, scale)
   - Output format (PNG or PDF)
   - Email recipients
   - Email subject and body

4. Save the job - it will be scheduled automatically

### Job Model

```json
{
  "id": "job-123",
  "cron": "0 9 * * *",
  "dashboardUid": "abc123",
  "slug": "my-dashboard",
  "panelId": 2,
  "from": "now-24h",
  "to": "now",
  "width": 1920,
  "height": 1080,
  "scale": 1,
  "format": "png",
  "recipients": ["user@example.com"],
  "subject": "Daily Report",
  "body": "Please find attached your daily report.",
  "variables": {
    "region": "us-east",
    "environment": "production"
  }
}
```

## API Endpoints

The plugin provides the following backend API endpoints:

- `GET /api/plugins/progressio-grafanareporter-app/resources/jobs` - List all jobs
- `POST /api/plugins/progressio-grafanareporter-app/resources/jobs` - Create a new job
- `GET /api/plugins/progressio-grafanareporter-app/resources/jobs/{id}` - Get job by ID
- `PUT /api/plugins/progressio-grafanareporter-app/resources/jobs/{id}` - Update job
- `DELETE /api/plugins/progressio-grafanareporter-app/resources/jobs/{id}` - Delete job
- `POST /api/plugins/progressio-grafanareporter-app/resources/jobs/{id}/execute` - Execute job immediately
- `GET /api/plugins/progressio-grafanareporter-app/resources/dashboards` - List all dashboards from Grafana
- `PUT /api/plugins/progressio-grafanareporter-app/resources/jobs/{id}` - Update job
- `DELETE /api/plugins/progressio-grafanareporter-app/resources/jobs/{id}` - Delete job
- `POST /api/plugins/progressio-grafanareporter-app/resources/jobs/{id}/execute` - Execute job immediately

## Architecture

### Backend (Go)

- **Scheduler**: Uses `robfig/cron` for cron-based scheduling
- **Job Storage**: Jobs are stored in `data/jobs.json`
- **Rendering**: Uses Grafana's `/render/d` and `/render/d-solo` endpoints
- **Email**: SMTP-based email delivery with attachments

### Frontend (React)

- **Job Management UI**: Create, edit, delete, and execute jobs
- **Dashboard Integration**: Seamless integration with Grafana UI
- **Real-time Updates**: Refresh job list and view status

## Development

```bash
# Run tests
go test ./...
npm test

# Lint
npm run lint
npm run lint:fix
```

## Environment Variables

- `GRAFANA_URL`: Grafana instance URL (default: `http://localhost:3000`)
- `SMTP_HOST`: SMTP server hostname
- `SMTP_PORT`: SMTP server port (default: `587`)
- `SMTP_USER`: SMTP username
- `SMTP_PASS`: SMTP password
- `SMTP_FROM`: From email address

## Troubleshooting

### Jobs not executing

1. Check Grafana logs for errors
2. Verify SMTP configuration
3. Ensure grafana-image-renderer is installed and working
4. Check API key permissions

### Email not sending

1. Verify SMTP settings
2. Check firewall rules
3. Test SMTP credentials manually

## License

Apache-2.0

## Contributing

Contributions are welcome! Please open an issue or pull request.
