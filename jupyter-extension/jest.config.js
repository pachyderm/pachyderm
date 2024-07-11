const baseConfig = require('@pachyderm/config/jest.config');

const esModules = [
  '@jupyterlab/',
  'lib0',
  'y\\-protocols',
  'y\\-websocket',
  'yjs',
].join('|');

baseConfig.transformIgnorePatterns = [`/node_modules/(?!${esModules}).+`];
baseConfig.preset = 'ts-jest/presets/js-with-babel';
baseConfig.testEnvironment = 'jsdom';
// TODO: Reenable this once we figure out why it is causing issues with stack traces being inaccurate and general display bugs.
// baseConfig.setupFilesAfterEnv.push('<rootDir>/scripts/jest.setup.js');

// On the current version of Jest we are using it will warn us about real timers being cleared when using `advanceTimersByTime`
// regardless of if we setup fake timers or not. In future versions of jest timers are rewritten again and should not have this issue.
baseConfig.timers = 'legacy';
module.exports = baseConfig;
