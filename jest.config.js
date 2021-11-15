const baseConfig = require('@pachyderm/config/jest.config');

const esModules = [
  '@jupyterlab/',
  'lib0',
  'y\\-protocols',
  'y\\-websocket',
  'yjs'
].join('|');

baseConfig.transformIgnorePatterns = [`/node_modules/(?!${esModules}).+`];
baseConfig.preset = 'ts-jest/presets/js-with-babel';

module.exports = baseConfig;
