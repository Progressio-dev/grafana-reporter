import type { Configuration } from 'webpack';
import path from 'path';
import CopyWebpackPlugin from 'copy-webpack-plugin';
import ForkTsCheckerWebpackPlugin from 'fork-ts-checker-webpack-plugin';
import ReplaceInFileWebpackPlugin from 'replace-in-file-webpack-plugin';
import LiveReloadPlugin from 'webpack-livereload-plugin';
const packageJson = require('../../package.json');

const DIST_DIR = path.resolve(__dirname, '../../dist');
const SRC_DIR = path.resolve(__dirname, '../../src');

const baseConfig = (env: any): Configuration => ({
  target: 'web',
  mode: env.production ? 'production' : 'development',
  devtool: env.production ? 'source-map' : 'eval-source-map',
  entry: {
    module: path.resolve(SRC_DIR, 'module.ts'),
  },
  output: {
    path: DIST_DIR,
    filename: '[name].js',
    libraryTarget: 'amd',
    clean: false,
  },
  externals: [
    'lodash',
    'react',
    'react-dom',
    '@grafana/data',
    '@grafana/ui',
    '@grafana/runtime',
    (context: any, request: any, callback: any) => {
      const prefix = 'grafana/';
      if (request && request.indexOf(prefix) === 0) {
        return callback(undefined, request.substr(prefix.length));
      }
      callback(undefined, undefined);
    },
  ],
  plugins: [
    new CopyWebpackPlugin({
      patterns: [
        { from: 'src/plugin.json', to: '.' },
        { from: 'src/img/*', to: 'img/[name][ext]', noErrorOnMissing: true },
        { from: 'README.md', to: '.', noErrorOnMissing: true },
        { from: 'CHANGELOG.md', to: '.', noErrorOnMissing: true },
        { from: 'LICENSE', to: '.', noErrorOnMissing: true },
      ],
    }),
    new ForkTsCheckerWebpackPlugin({
      async: Boolean(env.development),
      typescript: { configFile: path.resolve(__dirname, '../../tsconfig.json') },
    }),
    new ReplaceInFileWebpackPlugin([
      {
        dir: DIST_DIR,
        files: ['plugin.json'],
        rules: [
          {
            search: '%VERSION%',
            replace: packageJson.version,
          },
          {
            search: '%TODAY%',
            replace: new Date().toISOString().substring(0, 10),
          },
        ],
      },
    ]),
    ...(env.development ? [new LiveReloadPlugin()] : []),
  ],
  resolve: {
    extensions: ['.ts', '.tsx', '.js', '.jsx', '.json'],
    modules: [SRC_DIR, 'node_modules'],
  },
  module: {
    rules: [
      {
        test: /\.(ts|tsx)$/,
        exclude: /node_modules/,
        use: {
          loader: 'swc-loader',
          options: {
            jsc: {
              parser: {
                syntax: 'typescript',
                tsx: true,
                decorators: false,
                dynamicImport: true,
              },
            },
          },
        },
      },
      {
        test: /\.(scss|sass)$/,
        use: ['style-loader', 'css-loader', 'sass-loader'],
      },
      {
        test: /\.css$/,
        use: ['style-loader', 'css-loader'],
      },
      {
        test: /\.(png|jpe?g|gif|svg)$/,
        type: 'asset/resource',
      },
    ],
  },
});

export default baseConfig;
