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
    {
      displayName: 'test',
      ...baseConfig,
      testMatch: ['**/__tests__/**/*.test.ts'],
      setupFilesAfterEnv: [],
    },
  ],
};
