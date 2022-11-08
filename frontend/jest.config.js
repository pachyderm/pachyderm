/* eslint-disable @typescript-eslint/no-var-requires */

const {pathsToModuleNameMapper} = require('ts-jest/utils');

const tsConfig = require('./tsconfig.json');

const moduleNameMapper = pathsToModuleNameMapper(
  tsConfig.compilerOptions.paths,
  {
    prefix: '<rootDir>/src',
  },
);

module.exports = {
  cacheDirectory: '.jestcache',
  clearMocks: true,
  collectCoverageFrom: [
    '<rootDir>/src/**/*.{js,jsx,ts,tsx}',
    '<rootDir>/components/**/*.{js,jsx,ts,tsx}',
  ],
  coverageReporters: ['text-summary', 'lcov', 'json'],
  modulePaths: ['<rootDir>/src/'],
  moduleDirectories: ['src', 'node_modules'],
  roots: ['<rootDir>/src/', '<rootDir>/components/'],
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json', 'node'],
  setupFilesAfterEnv: ['<rootDir>/src/setupTests.ts'],
  testRegex: '(/__tests__/.*|(\\.|/)(test|spec))\\.tsx?$',
  testTimeout: 20000,
  timers: 'real',
  transform: {
    '^.+\\.tsx?$': '@swc/jest',
  },
  testEnvironment: './dash-jsdom-environment',
  moduleNameMapper: {
    '\\.(css)$': 'identity-obj-proxy',
    '\\.(gif|jpg|png)$': '<rootDir>/jest.file.mock.js',
    '\\.svg': '<rootDir>/jest.svg.mock.js',
    ...moduleNameMapper,
    d3: '<rootDir>/node_modules/d3/dist/d3.min.js',
    '^d3-(.*)$': `d3-$1/dist/d3-$1`,
  },
};
