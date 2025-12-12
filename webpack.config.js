const path = require('path');
const CopyWebpackPlugin = require('copy-webpack-plugin');
const ForkTsCheckerWebpackPlugin = require('fork-ts-checker-webpack-plugin');

const DIST_DIR = path.resolve(__dirname, 'dist');
const SRC_DIR = path.resolve(__dirname, 'src');

module.exports = (env) => ({
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
  ],
  plugins: [
    new CopyWebpackPlugin({
      patterns: [
        { from: 'src/plugin.json', to: '.' },
        { from: 'src/img', to: 'img', noErrorOnMissing: true },
        { from: 'README.md', to: '.', noErrorOnMissing: true },
        { from: 'CHANGELOG.md', to: '.', noErrorOnMissing: true },
        { from: 'LICENSE', to: '.', noErrorOnMissing: true },
      ],
    }),
    // Temporarily disabled - has type checking issues with Grafana UI components
    // new ForkTsCheckerWebpackPlugin({
    //   async: Boolean(env.development),
    //   typescript: {
    //     configFile: path.resolve(__dirname, 'tsconfig.json'),
    //     diagnosticOptions: {
    //       semantic: true,
    //       syntactic: false,
    //     },
    //   },
    // }),
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
