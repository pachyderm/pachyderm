/* eslint-disable @typescript-eslint/no-var-requires */

const {pathsToModuleNameMapper} = require('ts-jest/utils');

const baseConfig = require('../../jest.config.js');
const tsConfig = require('../tsconfig.json');

const moduleNameMapper = pathsToModuleNameMapper(
  tsConfig.compilerOptions.paths,
  {
    prefix: '<rootDir>/../src',
  },
);

baseConfig.moduleNameMapper = {
  ...baseConfig.moduleNameMapper,
  ...moduleNameMapper,
};

module.exports = baseConfig;
