/* eslint-disable @typescript-eslint/no-var-requires */

const {pathsToModuleNameMapper} = require('ts-jest');

const baseConfig = require('../../jest.config.base.js');
const tsConfig = require('../tsconfig.json');

const moduleNameMapper = pathsToModuleNameMapper(
  tsConfig.compilerOptions.paths,
  {
    prefix: '<rootDir>',
  },
);

baseConfig.testEnvironment = '../dash-jsdom-environment';
baseConfig.moduleNameMapper = {
  ...baseConfig.moduleNameMapper,
  ...moduleNameMapper,
};

process.env.TZ = 'UTC'; // Set timezone so tests with times pass locally and in CI.

module.exports = {
  ...baseConfig,
  testMatch: ['**/__tests__/**/*.test.tsx', '**/__tests__/**/*.test.ts'],
  testTimeout: baseConfig.testTimeout, // testTimeout doesn't get picked up inside of projects https://github.com/jestjs/jest/issues/9696
};
