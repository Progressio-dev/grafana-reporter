# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
