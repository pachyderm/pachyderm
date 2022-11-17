module.exports = {
  cacheDirectory: '.jestcache',
  clearMocks: true,
  collectCoverageFrom: ['<rootDir>/**/*.{js,jsx,ts,tsx}'],
  coverageReporters: ['text-summary', 'lcov', 'json'],
  modulePaths: ['<rootDir>'],
  moduleDirectories: ['src', 'node_modules'],
  roots: ['<rootDir>'],
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json', 'node'],
  setupFilesAfterEnv: ['./setupTests.ts'],
  testRegex: '(/__tests__/.*|(\\.|/)(test|spec))\\.tsx?$',
  testTimeout: 20000,
  timers: 'real',
  transform: {
    '^.+\\.tsx?$': '@swc/jest',
  },
  testEnvironment: 'jsdom',
};
