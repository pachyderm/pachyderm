const baseConfig = {
  cacheDirectory: '.jestcache',
  clearMocks: true,
  collectCoverageFrom: ['<rootDir>/**/*.{js,jsx,ts,tsx}'],
  modulePaths: ['<rootDir>'],
  moduleDirectories: ['src', 'node_modules'],
  roots: ['<rootDir>'],
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json', 'node'],
  moduleNameMapper: {
    '\\.(css)$': 'identity-obj-proxy',
    '\\.(gif|jpg|png)$': '<rootDir>/../jest.file.mock.js',
    '\\.svg': '<rootDir>/../jest.svg.mock.js',
    d3: '<rootDir>/../node_modules/d3/dist/d3.min.js',
    '^d3-(.*)$': `d3-$1/dist/d3-$1`,
  },
  setupFilesAfterEnv: ['./setupTests.ts'],
  fakeTimers: {},
  transform: {
    '^.+\\.tsx?$': '@swc/jest',
  },
  testEnvironment: 'jsdom',
};

// Jest does not support es modules yet
// https://github.com/remarkjs/react-markdown/issues/635#issuecomment-956158474
const esModules = [
  'react-markdown',
  'vfile',
  'unist-.+',
  'unified',
  'bail',
  'is-plain-obj',
  'trough',
  'remark-.+',
  'mdast-util-.+',
  'micromark',
  'parse-entities',
  'character-entities',
  'property-information',
  'comma-separated-tokens',
  'hast-util-whitespace',
  'space-separated-tokens',
  'decode-named-character-reference',
  'ccount',
  'escape-string-regexp',
  'markdown-table',
  'trim-lines',
  '@codemirror*',
  'json-schema-library',
].join('|');

baseConfig.transform = baseConfig.transform ?? {};
baseConfig.transform[`(${esModules}).+\\.js$`] = '@swc/jest';
baseConfig.transformIgnorePatterns = baseConfig.transformIgnorePatterns ?? [];
baseConfig.transformIgnorePatterns.push(
  `[/\\\\]node_modules[/\\\\](?!${esModules}).+\\.(js|jsx|mjs|cjs|ts|tsx)$`,
);

module.exports = baseConfig;
