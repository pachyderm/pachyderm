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
  '\\.(css)$': 'identity-obj-proxy',
  '\\.(gif|jpg|png)$': '<rootDir>/../jest.file.mock.js',
  '\\.svg': '<rootDir>/../jest.svg.mock.js',
  ...moduleNameMapper,
  d3: '<rootDir>/../node_modules/d3/dist/d3.min.js',
  '^d3-(.*)$': `d3-$1/dist/d3-$1`,
};

module.exports = baseConfig;
