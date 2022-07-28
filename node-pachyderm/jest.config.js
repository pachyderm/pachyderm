/* eslint-disable @typescript-eslint/no-var-requires */
const baseConfig = require('@pachyderm/config/jest.config');

baseConfig.testEnvironment = 'node';
baseConfig.timers = 'real';

module.exports = baseConfig;
