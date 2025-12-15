import React, { useState, useEffect } from 'react';
import { Select, Field, Alert } from '@grafana/ui';
import { SelectableValue } from '@grafana/data';
import { getBackendSrv } from '@grafana/runtime';

interface Dashboard {
  id: number;
  uid: string;
  title: string;
  uri: string;
  url: string;
  slug: string;
  type: string;
  tags: string[];
  isStarred: boolean;
  folderId?: number;
  folderUid?: string;
  folderTitle?: string;
  folderUrl?: string;
}

interface DashboardSelectorProps {
  pluginId: string;
  value: { uid: string; slug: string } | null;
  onChange: (dashboard: { uid: string; slug: string; title: string } | null) => void;
}

export const DashboardSelector: React.FC<DashboardSelectorProps> = ({ pluginId, value, onChange }) => {
  const [dashboards, setDashboards] = useState<Dashboard[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadDashboards();
  }, []);

  const loadDashboards = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await getBackendSrv().get(`/api/plugins/${pluginId}/resources/dashboards`);
      setDashboards(response || []);
    } catch (err) {
      console.error('Failed to load dashboards:', err);
      setError('Failed to load dashboards from Grafana. Make sure the plugin is configured correctly.');
    } finally {
      setLoading(false);
    }
  };

  // Group dashboards by folder
  const groupedOptions: Array<SelectableValue<string>> = [];
  const folderGroups: { [key: string]: Dashboard[] } = {};

  dashboards.forEach((dashboard) => {
    const folderTitle = (dashboard.folderTitle?.trim() || '');
    const effectiveFolderTitle = folderTitle === '' ? 'General' : folderTitle;
    if (!folderGroups[effectiveFolderTitle]) {
      folderGroups[effectiveFolderTitle] = [];
    }
    folderGroups[effectiveFolderTitle].push(dashboard);
  });

  // Sort folders alphabetically, but keep "General" first
  const sortedFolders = Object.keys(folderGroups).sort((a, b) => {
    if (a === 'General') return -1;
    if (b === 'General') return 1;
    return a.localeCompare(b);
  });

  sortedFolders.forEach((folderTitle) => {
    const dashboardsInFolder = folderGroups[folderTitle];
    
    // Sort dashboards alphabetically within each folder
    dashboardsInFolder.sort((a, b) => a.title.localeCompare(b.title));
    
    dashboardsInFolder.forEach((dashboard) => {
      const label = folderTitle === 'General' 
        ? dashboard.title 
        : `${folderTitle} / ${dashboard.title}`;
      
      groupedOptions.push({
        label,
        value: dashboard.uid,
        description: dashboard.uid,
      });
    });
  });

  const handleChange = (option: SelectableValue<string>) => {
    if (!option.value) {
      onChange(null);
      return;
    }

    const selectedDashboard = dashboards.find((d) => d.uid === option.value);
    if (selectedDashboard) {
      onChange({
        uid: selectedDashboard.uid,
        slug: selectedDashboard.slug,
        title: selectedDashboard.title,
      });
    }
  };

  const selectedOption = value
    ? groupedOptions.find((opt) => opt.value === value.uid)
    : null;

  return (
    <div>
      {error && (
        <Alert title="Error" severity="error" style={{ marginBottom: '10px' }}>
          {error}
        </Alert>
      )}
      <Field label="Dashboard" description="Select a dashboard from your Grafana instance">
        <Select
          options={groupedOptions}
          value={selectedOption}
          onChange={handleChange}
          isLoading={loading}
          placeholder="Select a dashboard..."
          isClearable
        />
      </Field>
    </div>
  );
};
