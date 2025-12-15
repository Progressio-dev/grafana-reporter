# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2025-12-15

### Added
- Plugin version display showing version, build time, start time, and uptime on the main UI
- Force reload functionality to reload plugin configuration and jobs without restarting Grafana
- `/version` API endpoint to retrieve plugin version and runtime information
- `/reload` API endpoint to force reload plugin configuration and scheduled jobs
- Build-time version injection via Makefile using git tags and timestamps
- Enhanced dashboard slug extraction to properly handle all dashboard URI formats

### Changed
- Makefile now injects version and build time during compilation
- Dashboard API now processes and ensures all dashboards have proper slugs
- Version information is automatically displayed at the top of the plugin UI

### Fixed
- Dashboard selector now properly extracts slugs from dashboard URIs and URLs
- Improved dashboard selection reliability by handling missing slug fields
- Fallback to UID when slug cannot be determined from URI/URL

## [1.1.0] - 2025-12-15

### Added
- HTML email format option that embeds dashboard images directly in the email body
- Improved dashboard selector to properly display root/General folder dashboards

### Changed
- Plugin version incremented from 1.0.2 to 1.1.0 to ensure Grafana recognizes the update
- Enhanced email sending to support multipart/related HTML emails with embedded images

### Fixed
- Dashboard selector now correctly handles dashboards without folder assignments (root/General folder)
- Empty folder titles are now properly normalized to "General"

## [1.0.0] - 2025-12-12

### Added
- Initial release of Grafana Reporter Plugin
- Cron-based scheduler for automated report generation
- Support for PDF and PNG report formats
- Dashboard and panel rendering via Grafana's render API
- Email delivery with SMTP support
- Job management UI (create, edit, delete, execute)
- REST API for job management
- JSON-based job storage
- Health check endpoint
- Docker support with docker-compose for local development
- CI/CD pipeline with GitHub Actions

### Features
- **Backend (Go)**:
  - Plugin implementation using grafana-plugin-sdk-go
  - Cron scheduler using robfig/cron
  - Grafana render API client
  - SMTP email sender with attachment support
  - Resource handlers for CRUD operations on jobs
  - Health check endpoint

- **Frontend (React)**:
  - Job management interface
  - Cron expression configuration
  - Dashboard and panel selection
  - Email configuration (recipients, subject, body)
  - Time range selection (from/to)
  - Render settings (width, height, scale, format)
  - Manual job execution

### Requirements
- Grafana 9.0+
- grafana-image-renderer plugin
- SMTP server for email delivery

### Configuration
- Environment variables for SMTP settings
- Plugin configuration for Grafana API access
- JSON file-based job storage
