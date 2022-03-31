/* eslint-disable @typescript-eslint/no-var-requires */


const baseConfig = require('@pachyderm/config/jest.config');
const {pathsToModuleNameMapper} = require('ts-jest');

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
  d3: '<rootDir>/node_modules/d3/dist/d3.min.js',
  '^d3-(.*)$': `d3-$1/dist/d3-$1`,
};

baseConfig.timers = 'real';
baseConfig.testEnvironment = './dash-jsdom-environment';

module.exports = baseConfig;
