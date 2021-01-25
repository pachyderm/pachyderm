const fs = require('fs');
const path = require('path');

const config = require('@pachyderm/config/webpack.config'); 
const TsconfigPathsPlugin = require('tsconfig-paths-webpack-plugin');

const appDir = fs.realpathSync(process.cwd());


config.module.rules[0].oneOf[1].include.push(path.resolve(appDir, '..', 'backend', 'src'))

module.exports = {
  ...config,
  resolve: {
    ...config.resolve,
    plugins: [
      new TsconfigPathsPlugin(),
    ],
  }
}
