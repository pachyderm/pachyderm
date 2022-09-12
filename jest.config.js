module.exports = {
  cacheDirectory: '.jestcache',
  clearMocks: true,
  collectCoverageFrom: [
    '<rootDir>/src/**/*.{js,jsx,ts,tsx}',
  ],
  coverageReporters: ['text-summary', 'lcov', 'json'],
  modulePaths: ['<rootDir>/src/'],
  moduleDirectories: ['src', 'node_modules'],
  roots: ['<rootDir>/src/'],
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json', 'node'],
  setupFilesAfterEnv: ['<rootDir>/src/setupTests.ts'],
  testRegex: '(/__tests__/.*|(\\.|/)(test|spec))\\.tsx?$',
  testTimeout: 20000,
  timers: 'modern',
  transform: {
    '^.+\\.tsx?$': 'ts-jest',
  },
  transformIgnorePatterns: [
    "node_modules/(?!(@pachyderm/components)/)"
  ],
};
