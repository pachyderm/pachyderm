/* eslint-disable @typescript-eslint/no-var-requires */
const {pathsToModuleNameMapper} = require('ts-jest');

const baseConfig = require('../jest.config.base.js');

const tsConfig = require('./tsconfig.json');

const moduleNameMapper = pathsToModuleNameMapper(
  tsConfig.compilerOptions.paths,
  {
    prefix: '<rootDir>/src',
  },
);

baseConfig.moduleNameMapper = {
  ...moduleNameMapper,
};

baseConfig.testEnvironment = 'node';

module.exports = {
  projects: [
    // TODO: add a no-mock test type.
    {
      displayName: 'mock-server',
      ...baseConfig,
      testMatch: [
        '**/__tests__/**/*.test.ts',
        '!**/__tests__/**/*.network-mock.test.ts',
      ],
      setupFilesAfterEnv: ['./setupTests.ts'],
    },
    {
      displayName: 'network-mock-test',
      ...baseConfig,
      testMatch: ['**/__tests__/**/*.network-mock.test.ts'],
      setupFilesAfterEnv: ['./setupTestsNetworkMock.ts'],
    },
    {
      displayName: 'integration-test',
      ...baseConfig,
      testMatch: ['**/__tests__/**/*.integration.ts'],
      setupFilesAfterEnv: [],
    },
    {
      displayName: 'unit-test',
      ...baseConfig,
      testMatch: ['**/__tests__/**/*.unit.ts'],
      setupFilesAfterEnv: [],
    },
  ],
};
