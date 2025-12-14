import React from 'react';
import { Button, Icon } from '@grafana/ui';

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
  variables?: { [key: string]: string };
}

interface JobListProps {
  jobs: Job[];
  onEdit: (job: Job) => void;
  onDelete: (jobId: string) => void;
  onExecute: (jobId: string) => void;
}

export const JobList: React.FC<JobListProps> = ({ jobs, onEdit, onDelete, onExecute }) => {
  if (jobs.length === 0) {
    return (
      <div style={{ padding: '20px', textAlign: 'center', color: '#999' }}>
        No jobs configured. Click "Create Job" to add your first scheduled report.
      </div>
    );
  }

  return (
    <div>
      <table style={{ width: '100%', borderCollapse: 'collapse' }}>
        <thead>
          <tr style={{ borderBottom: '2px solid #333' }}>
            <th style={{ padding: '10px', textAlign: 'left' }}>Dashboard</th>
            <th style={{ padding: '10px', textAlign: 'left' }}>Schedule</th>
            <th style={{ padding: '10px', textAlign: 'left' }}>Format</th>
            <th style={{ padding: '10px', textAlign: 'left' }}>Recipients</th>
            <th style={{ padding: '10px', textAlign: 'right' }}>Actions</th>
          </tr>
        </thead>
        <tbody>
          {jobs.map((job) => (
            <tr key={job.id} style={{ borderBottom: '1px solid #444' }}>
              <td style={{ padding: '10px' }}>
                <div>
                  <strong>{job.slug}</strong>
                </div>
                <div style={{ fontSize: '0.9em', color: '#999' }}>
                  {job.dashboardUid}
                  {job.panelId && ` (Panel ${job.panelId})`}
                </div>
              </td>
              <td style={{ padding: '10px' }}>
                <code style={{ fontSize: '0.9em' }}>{job.cron}</code>
              </td>
              <td style={{ padding: '10px' }}>
                <span style={{ textTransform: 'uppercase' }}>{job.format}</span>
                <div style={{ fontSize: '0.9em', color: '#999' }}>
                  {job.width}x{job.height} @{job.scale}x
                </div>
              </td>
              <td style={{ padding: '10px' }}>
                <div style={{ fontSize: '0.9em' }}>
                  {job.recipients.length} recipient(s)
                </div>
              </td>
              <td style={{ padding: '10px', textAlign: 'right' }}>
                <Button
                  size="sm"
                  variant="secondary"
                  icon="play"
                  onClick={() => onExecute(job.id)}
                  style={{ marginRight: '5px' }}
                  title="Execute now"
                />
                <Button
                  size="sm"
                  variant="secondary"
                  icon="edit"
                  onClick={() => onEdit(job)}
                  style={{ marginRight: '5px' }}
                  title="Edit"
                />
                <Button
                  size="sm"
                  variant="destructive"
                  icon="trash-alt"
                  onClick={() => onDelete(job.id)}
                  title="Delete"
                />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
