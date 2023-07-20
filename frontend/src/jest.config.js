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

module.exports = {
  testTimeout: baseConfig.testTimeout, // testTimeout doesn't get picked up inside of projects https://github.com/jestjs/jest/issues/9696
  projects: [
    {
      displayName: 'mock-server',
      ...baseConfig,
      testMatch: ['**/__tests__/**/*.test.mock.tsx'],
      setupFilesAfterEnv: ['./setupTests.mock.ts'],
    },
    {
      displayName: 'msw',
      ...baseConfig,
      testMatch: ['**/__tests__/**/*.test.tsx'],
    },
    {
      displayName: 'unit',
      ...baseConfig,
      testMatch: ['**/__tests__/**/*.test.ts', '**/__tests__/**/*.unit.ts'],
    },
  ],
};
