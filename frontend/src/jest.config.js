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
  testTimeout: baseConfig.testTimeout, // testTimeout doesn't get picked up inside of projects https://github.com/jestjs/jest/issues/9696
  projects: [
    {
      ...baseConfig,
      displayName: 'mock-server',
      testMatch: ['**/__tests__/**/*.mock.test.ts?(x)'],
      setupFilesAfterEnv: ['./setupTests.mock.ts'],
    },
    {
      ...baseConfig,
      displayName: 'msw',
      testMatch: [
        '**/__tests__/**/*.test.tsx',
        '!**/__tests__/**/*.mock.test.ts?(x)',
        '!**/__tests__/**/*.unit.test.ts?(x)',
      ],
    },
    {
      ...baseConfig,
      displayName: 'unit',
      testMatch: ['**/__tests__/**/*.unit.test.ts?(x)'],
    },
  ],
};
