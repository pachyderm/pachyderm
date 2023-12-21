const glob = require('glob');
const path = require('path');

module.exports = {
  mode: 'production',
  entry: glob.sync('./tests/*.ts').reduce((acc, match) => {
    const entry = path.basename(match, '.ts');
    acc[entry] = './' + match;
    return acc;
  }, {}),
  output: {
    path: path.resolve(__dirname, 'build'),
    libraryTarget: 'commonjs',
    filename: '[name].bundle.js',
  },
  resolve: {
    extensions: ['.ts', '.js'],
    alias: {
      '@queries': path.resolve(__dirname, '../frontend/src/queries'),
      '@mutations': path.resolve(__dirname, '../frontend/src/mutations'),
      '@dash-frontend': path.resolve(__dirname, '../frontend/src/'),
    },
  },
  module: {
    rules: [{test: /\.ts$/, use: 'babel-loader'}],
  },
  target: 'web',
  externals: /k6(\/.*)?/,
};
