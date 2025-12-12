import type { Configuration } from 'webpack';
import grafanaConfig from './.config/webpack/webpack.config';

const config = async (env: any): Promise<Configuration> => {
  return grafanaConfig(env);
};

export default config;
