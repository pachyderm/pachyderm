/* eslint-disable @typescript-eslint/no-var-requires */
const {pathsToModuleNameMapper} = require('ts-jest');

const baseConfig = require('../jest.config.js');

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
baseConfig.timers = 'real';

module.exports = baseConfig;
