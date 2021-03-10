module.exports = {
  cacheDirectory: '.jestcache',
  clearMocks: true,
  collectCoverageFrom: [
    '<rootDir>/src/**/*.{js,jsx,ts,tsx}',
    '!<rootDir>/src/components/Svg/**/*',
    '!<rootDir>/src/**/*.stories.{js,jsx,ts,tsx}',
    '!<rootDir>/src/mock/**/*',
  ],
  coverageReporters: ['text-summary', 'lcov', 'json'],
  modulePaths: ['<rootDir>/src/'],
  moduleDirectories: ['src', 'node_modules'],
  roots: ['<rootDir>/src/'],
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json', 'node'],
  moduleNameMapper: {
    '\\.(css)$': 'identity-obj-proxy',
    '\\.(gif|jpg|png)$': '@pachyderm/config/jest.file.mock.js',
    '\\.svg': '@pachyderm/config/jest.svg.mock.js'
  },
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
