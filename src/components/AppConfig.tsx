import React, { useState, useEffect } from 'react';
import { AppRootProps, PluginConfigPageProps } from '@grafana/data';
import { Button, Field, Input, VerticalGroup, HorizontalGroup, Alert, Spinner } from '@grafana/ui';
import { getBackendSrv } from '@grafana/runtime';
import { JobList } from './JobList';
import { JobForm } from './JobForm';
import { Settings } from './Settings';

interface Job {
  id: string;
  cron: string;
  dashboardUid: string;
  slug: string;
  panelId?: number;
  from: string;
  to: string;
  width: number;
  height: number;
  scale: number;
  format: string;
  recipients: string[];
  subject: string;
  body: string;
}

export const AppConfig: React.FC<AppRootProps> = ({ meta, query }) => {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [showSettings, setShowSettings] = useState(false);
  const [editingJob, setEditingJob] = useState<Job | null>(null);

  const pluginId = meta.id;

  useEffect(() => {
    loadJobs();
  }, []);

  const loadJobs = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await getBackendSrv().get(`/api/plugins/${pluginId}/resources/jobs`);
      setJobs(response || []);
    } catch (err) {
      console.error('Failed to load jobs:', err);
      setError('Failed to load jobs. Make sure the plugin backend is running.');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateJob = () => {
    setEditingJob(null);
    setShowForm(true);
  };

  const handleEditJob = (job: Job) => {
    setEditingJob(job);
    setShowForm(true);
  };

  const handleDeleteJob = async (jobId: string) => {
    // TODO: Replace with Grafana UI confirmation dialog component
    if (!window.confirm('Are you sure you want to delete this job?')) {
      return;
    }

    try {
      await getBackendSrv().delete(`/api/plugins/${pluginId}/resources/jobs/${jobId}`);
      await loadJobs();
    } catch (err) {
      console.error('Failed to delete job:', err);
      // TODO: Replace with Grafana UI notification system
      window.alert('Failed to delete job');
    }
  };

  const handleExecuteJob = async (jobId: string) => {
    try {
      await getBackendSrv().post(`/api/plugins/${pluginId}/resources/jobs/${jobId}/execute`);
      // TODO: Replace with Grafana UI notification system
      window.alert('Job execution started');
    } catch (err) {
      console.error('Failed to execute job:', err);
      // TODO: Replace with Grafana UI notification system
      window.alert('Failed to execute job');
    }
  };

  const handleSaveJob = async (job: Job) => {
    try {
      if (editingJob) {
        // Update existing job
        await getBackendSrv().put(`/api/plugins/${pluginId}/resources/jobs/${job.id}`, job);
      } else {
        // Create new job
        await getBackendSrv().post(`/api/plugins/${pluginId}/resources/jobs`, job);
      }
      await loadJobs();
      setShowForm(false);
      setEditingJob(null);
    } catch (err) {
      console.error('Failed to save job:', err);
      throw err;
    }
  };

  const handleCancelForm = () => {
    setShowForm(false);
    setEditingJob(null);
  };

  const handleShowSettings = () => {
    setShowSettings(true);
  };

  const handleBackFromSettings = () => {
    setShowSettings(false);
  };

  if (loading) {
    return (
      <div style={{ padding: '20px' }}>
        <Spinner />
      </div>
    );
  }

  if (showSettings) {
    return <Settings pluginId={pluginId} onBack={handleBackFromSettings} />;
  }

  if (showForm) {
    return (
      <div style={{ padding: '20px' }}>
        <JobForm
          job={editingJob}
          onSave={handleSaveJob}
          onCancel={handleCancelForm}
        />
      </div>
    );
  }

  return (
    <div style={{ padding: '20px' }}>
      <VerticalGroup spacing="lg">
        <div>
          <h2>Grafana Reporter</h2>
          <p>Schedule PDF/PNG report generation and email sending</p>
        </div>

        {error && (
          <Alert title="Error" severity="error">
            {error}
          </Alert>
        )}

        <HorizontalGroup>
          <Button icon="plus" onClick={handleCreateJob}>
            Create Job
          </Button>
          <Button icon="cog" variant="secondary" onClick={handleShowSettings}>
            Settings
          </Button>
          <Button icon="sync" variant="secondary" onClick={loadJobs}>
            Refresh
          </Button>
        </HorizontalGroup>

        <JobList
          jobs={jobs}
          onEdit={handleEditJob}
          onDelete={handleDeleteJob}
          onExecute={handleExecuteJob}
        />
      </VerticalGroup>
    </div>
  );
};
