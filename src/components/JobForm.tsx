import React, { useState, useEffect } from 'react';
import {
  Button,
  Field,
  Input,
  TextArea,
  Select,
  VerticalGroup,
  HorizontalGroup,
  Alert,
} from '@grafana/ui';
import { SelectableValue } from '@grafana/data';

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

interface JobFormProps {
  job: Job | null;
  onSave: (job: Job) => Promise<void>;
  onCancel: () => void;
}

const formatOptions: Array<SelectableValue<string>> = [
  { label: 'PNG', value: 'png' },
  { label: 'PDF', value: 'pdf' },
];

export const JobForm: React.FC<JobFormProps> = ({ job, onSave, onCancel }) => {
  const [formData, setFormData] = useState<Job>({
    id: '',
    cron: '0 9 * * *',
    dashboardUid: '',
    slug: '',
    panelId: undefined,    
    from: 'now-24h',
    to: 'now',
    width: 1920,
    height: 1080,
    scale: 1,
    format: 'png',
    recipients: [],
    subject: 'Grafana Report',
    body: 'Please find attached your scheduled Grafana report.',
  });

  const [recipientsText, setRecipientsText] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (job) {
      setFormData(job);
      setRecipientsText(job.recipients.join(', '));
    }
  }, [job]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    // Validate form
    if (!formData.cron) {
      setError('Cron expression is required');
      return;
    }
    if (!formData.dashboardUid) {
      setError('Dashboard UID is required');
      return;
    }
    if (!formData.slug) {
      setError('Dashboard slug is required');
      return;
    }
    if (recipientsText.trim() === '') {
      setError('At least one recipient is required');
      return;
    }

    // Parse recipients
    const recipients = recipientsText
      .split(',')
      .map((email) => email.trim())
      .filter((email) => email.length > 0);

    if (recipients.length === 0) {
      setError('At least one recipient is required');
      return;
    }

    try {
      setSaving(true);
      await onSave({
        ...formData,
        recipients,
      });
    } catch (err) {
      setError('Failed to save job: ' + (err as Error).message);
    } finally {
      setSaving(false);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <VerticalGroup spacing="lg">
        <h3>{job ? 'Edit Job' : 'Create Job'}</h3>

        {error && (
          <Alert title="Error" severity="error">
            {error}
          </Alert>
        )}

        <Field label="Cron Expression" description="When to run the job (e.g., '0 9 * * *' for daily at 9 AM)">
          <Input
            value={formData.cron}
            onChange={(e) => setFormData({ ...formData, cron: e.currentTarget.value })}
            placeholder="0 9 * * *"
            required
          />
        </Field>

        <Field label="Dashboard UID" description="The UID of the dashboard to render">
          <Input
            value={formData.dashboardUid}
            onChange={(e) => setFormData({ ...formData, dashboardUid: e.currentTarget.value })}
            placeholder="abc123xyz"
            required
          />
        </Field>

        <Field label="Dashboard Slug" description="The URL slug of the dashboard">
          <Input
            value={formData.slug}
            onChange={(e) => setFormData({ ...formData, slug: e.currentTarget.value })}
            placeholder="my-dashboard"
            required
          />
        </Field>

        <Field label="Panel ID (optional)" description="Leave empty to render the full dashboard">
          <Input
            type="number"
	    value={formData.panelId !== undefined ?  formData.panelId : ''}
            onChange={(e) =>
              setFormData({
                ...formData,
                panelId: e.currentTarget.value ? parseInt(e.currentTarget.value) : undefined,
              })
            }
            placeholder="2"
          />
        </Field>

        <HorizontalGroup>
          <Field label="From">
            <Input
              value={formData.from}
              onChange={(e) => setFormData({ ...formData, from: e.currentTarget.value })}
              placeholder="now-24h"
              required
            />
          </Field>

          <Field label="To">
            <Input
              value={formData.to}
              onChange={(e) => setFormData({ ...formData, to: e.currentTarget.value })}
              placeholder="now"
              required
            />
          </Field>
        </HorizontalGroup>

        <HorizontalGroup>
          <Field label="Width">
            <Input
              type="number"
              value={formData.width}
              onChange={(e) => setFormData({ ...formData, width: parseInt(e.currentTarget.value) })}
              required
            />
          </Field>

          <Field label="Height">
            <Input
              type="number"
              value={formData.height}
              onChange={(e) => setFormData({ ...formData, height: parseInt(e.currentTarget.value) })}
              required
            />
          </Field>

          <Field label="Scale">
            <Input
              type="number"
              value={formData.scale}
              onChange={(e) => setFormData({ ...formData, scale: parseInt(e.currentTarget.value) })}
              required
            />
          </Field>
        </HorizontalGroup>

        <Field label="Format">
          <Select
            options={formatOptions}
            value={formData.format}
            onChange={(option) => setFormData({ ...formData, format: option.value! })}
          />
        </Field>

        <Field label="Recipients" description="Comma-separated list of email addresses">
          <Input
            value={recipientsText}
            onChange={(e) => setRecipientsText(e.currentTarget.value)}
            placeholder="user1@example.com, user2@example.com"
            required
          />
        </Field>

        <Field label="Email Subject">
          <Input
            value={formData.subject}
            onChange={(e) => setFormData({ ...formData, subject: e.currentTarget.value })}
            required
          />
        </Field>

        <Field label="Email Body">
          <TextArea
            value={formData.body}
            onChange={(e) => setFormData({ ...formData, body: e.currentTarget.value })}
            rows={5}
            required
          />
        </Field>

        <HorizontalGroup>
          <Button type="submit" disabled={saving}>
            {saving ? 'Saving...' : 'Save'}
          </Button>
          <Button variant="secondary" onClick={onCancel} disabled={saving}>
            Cancel
          </Button>
        </HorizontalGroup>
      </VerticalGroup>
    </form>
  );
};
