/* eslint-disable @typescript-eslint/no-var-requires */
const baseConfig = require('@pachyderm/config/jest.config');
const {pathsToModuleNameMapper} = require('ts-jest/utils');

const tsConfig = require('./tsconfig.json');

const moduleNameMapper = pathsToModuleNameMapper(
  tsConfig.compilerOptions.paths,
  {
    prefix: '<rootDir>/src',
  },
);

baseConfig.moduleNameMapper = {
  ...baseConfig.moduleNameMapper,
  ...moduleNameMapper,
};

module.exports = baseConfig;
