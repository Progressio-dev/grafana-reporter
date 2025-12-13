import React, { useState, useEffect } from 'react';
import {
  Button,
  Field,
  Input,
  VerticalGroup,
  HorizontalGroup,
  Alert,
  Spinner,
} from '@grafana/ui';
import { getBackendSrv } from '@grafana/runtime';

interface Config {
  grafanaUrl: string;
  grafanaApiKey: string;
  smtpHost: string;
  smtpPort: number;
  smtpUser: string;
  smtpPassword: string;
  smtpFrom: string;
}

interface SettingsProps {
  pluginId: string;
  onBack: () => void;
}

export const Settings: React.FC<SettingsProps> = ({ pluginId, onBack }) => {
  const [config, setConfig] = useState<Config>({
    grafanaUrl: 'http://localhost:3000',
    grafanaApiKey: '',
    smtpHost: '',
    smtpPort: 587,
    smtpUser: '',
    smtpPassword: '',
    smtpFrom: '',
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await getBackendSrv().get(`/api/plugins/${pluginId}/resources/config`);
      setConfig(response);
    } catch (err) {
      console.error('Failed to load configuration:', err);
      setError('Failed to load configuration');
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    try {
      setSaving(true);
      setError(null);
      setSuccess(null);
      await getBackendSrv().post(`/api/plugins/${pluginId}/resources/config`, config);
      setSuccess('Configuration saved successfully');
      // Reload to get masked values
      await loadConfig();
    } catch (err) {
      console.error('Failed to save configuration:', err);
      setError('Failed to save configuration');
    } finally {
      setSaving(false);
    }
  };

  const handleTestSMTP = async () => {
    // Validate before setting loading state
    if (!config.smtpFrom) {
      setError('Please configure "From Email" before testing');
      return;
    }

    try {
      setTesting(true);
      setError(null);
      setSuccess(null);

      await getBackendSrv().post(`/api/plugins/${pluginId}/resources/test-email`, {
        recipients: [config.smtpFrom],
        subject: 'Grafana Reporter Test Email',
        body: 'This is a test email from Grafana Reporter plugin. If you receive this, your SMTP configuration is working correctly.',
      });
      setSuccess('Test email sent successfully! Check your inbox.');
    } catch (err) {
      console.error('Failed to send test email:', err);
      setError('Failed to send test email. Please check your SMTP configuration.');
    } finally {
      setTesting(false);
    }
  };

  if (loading) {
    return (
      <div style={{ padding: '20px' }}>
        <Spinner />
      </div>
    );
  }

  return (
    <div style={{ padding: '20px', maxWidth: '800px' }}>
      <VerticalGroup spacing="lg">
        <div>
          <h2>Settings</h2>
          <p>Configure Grafana and SMTP settings for the reporter plugin</p>
        </div>

        {error && (
          <Alert title="Error" severity="error" onRemove={() => setError(null)}>
            {error}
          </Alert>
        )}

        {success && (
          <Alert title="Success" severity="success" onRemove={() => setSuccess(null)}>
            {success}
          </Alert>
        )}

        {/* Grafana Configuration */}
        <div>
          <h3>Grafana Configuration</h3>
          <VerticalGroup spacing="md">
            <Field label="Grafana URL" description="The URL of your Grafana instance">
              <Input
                value={config.grafanaUrl}
                onChange={(e) => setConfig({ ...config, grafanaUrl: e.currentTarget.value })}
                placeholder="http://localhost:3000"
              />
            </Field>
            <Field
              label="Grafana API Key"
              description="Service account token with appropriate permissions"
            >
              <Input
                type="password"
                value={config.grafanaApiKey}
                onChange={(e) => setConfig({ ...config, grafanaApiKey: e.currentTarget.value })}
                placeholder="Enter API key"
              />
            </Field>
          </VerticalGroup>
        </div>

        {/* SMTP Configuration */}
        <div>
          <h3>SMTP Configuration</h3>
          <VerticalGroup spacing="md">
            <Field label="SMTP Host" description="SMTP server hostname">
              <Input
                value={config.smtpHost}
                onChange={(e) => setConfig({ ...config, smtpHost: e.currentTarget.value })}
                placeholder="smtp.gmail.com"
              />
            </Field>
            <Field label="SMTP Port" description="SMTP server port (default: 587)">
              <Input
                type="number"
                value={config.smtpPort}
                onChange={(e) => {
                  const port = parseInt(e.currentTarget.value, 10);
                  if (isNaN(port) || port < 1 || port > 65535) {
                    setConfig({ ...config, smtpPort: 587 });
                  } else {
                    setConfig({ ...config, smtpPort: port });
                  }
                }}
                placeholder="587"
                min="1"
                max="65535"
              />
            </Field>
            <Field label="SMTP Username" description="Username for SMTP authentication">
              <Input
                value={config.smtpUser}
                onChange={(e) => setConfig({ ...config, smtpUser: e.currentTarget.value })}
                placeholder="username@example.com"
              />
            </Field>
            <Field label="SMTP Password" description="Password for SMTP authentication">
              <Input
                type="password"
                value={config.smtpPassword}
                onChange={(e) => setConfig({ ...config, smtpPassword: e.currentTarget.value })}
                placeholder="Enter password"
              />
            </Field>
            <Field label="From Email" description="Email address to use as sender">
              <Input
                type="email"
                value={config.smtpFrom}
                onChange={(e) => setConfig({ ...config, smtpFrom: e.currentTarget.value })}
                placeholder="noreply@example.com"
              />
            </Field>
          </VerticalGroup>
        </div>

        {/* Action Buttons */}
        <HorizontalGroup>
          <Button onClick={handleSave} disabled={saving}>
            {saving ? 'Saving...' : 'Save Configuration'}
          </Button>
          <Button variant="secondary" onClick={handleTestSMTP} disabled={testing}>
            {testing ? 'Testing...' : 'Test SMTP'}
          </Button>
          <Button variant="secondary" onClick={onBack}>
            Back to Jobs
          </Button>
        </HorizontalGroup>
      </VerticalGroup>
    </div>
  );
};
