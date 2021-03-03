/* eslint-disable @typescript-eslint/no-var-requires */
const fs = require('fs');
const path = require('path');

const config = require('@pachyderm/config/webpack.config');
const TsconfigPathsPlugin = require('tsconfig-paths-webpack-plugin');
const {DefinePlugin} = require('webpack');

const NODE_ENV = process.env.NODE_ENV;

const appDir = fs.realpathSync(process.cwd());
const isProd = NODE_ENV === 'production' || NODE_ENV === 'staging';

const env = Object.keys(process.env)
  .filter((key) => key.startsWith('REACT_APP') || key === 'NODE_ENV')
  .reduce((env, key) => {
    if (key.startsWith('REACT_APP_RUNTIME')) {
      env.pachDashConfig[key] = isProd ? `window.pachDashConfig['${key}']` : (JSON.stringify(process.env[key]) || '');
    } else {
      env[key] = JSON.stringify(process.env[key]) || '';
    }

    return env;
  }, { pachDashConfig: {} });

config.module.rules[0].oneOf[1].include.push(
  path.resolve(appDir, '..', 'backend', 'src'),
);

module.exports = {
  ...config,
  resolve: {
    ...config.resolve,
    plugins: [new TsconfigPathsPlugin()],
  },
  plugins: [
    new DefinePlugin({'process.env': env}),
    ...config.plugins,
  ].filter(Boolean),
};
