/* eslint-disable @typescript-eslint/no-var-requires */
const path = require('path');

const config = {
  mode: 'production',
  devtool: 'hidden-source-map',
  entry: './index.js',
  output: {
    path: path.resolve(__dirname, 'dist'),
    filename: './index.js',
    libraryTarget: 'umd',
    globalObject: 'this',
  },
  module: {
    rules: [
      {
        oneOf: [
          {
            test: /\.(js)$/,
            loader: require.resolve('babel-loader'),
            options: {
              cacheDirectory: true,
              cacheCompression: false,
              compact: true,
            },
          },
        ],
      },
    ],
  },
};

module.exports = config;
