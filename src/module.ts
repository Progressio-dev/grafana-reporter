import { AppPlugin } from '@grafana/data';
import { AppConfig } from './components/AppConfig';

export const plugin = new AppPlugin().setRootPage(AppConfig);
