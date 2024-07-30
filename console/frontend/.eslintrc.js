/* eslint-disable @typescript-eslint/no-var-requires */

module.exports = {
  root: true,
  extends: [
    'react-app',
    'eslint:recommended',
    'plugin:react/recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:prettier/recommended',
    'plugin:import/errors',
    'plugin:import/warnings',
    'plugin:import/typescript',
    'plugin:jest/recommended',
    'plugin:jest/style',
    'plugin:jest-dom/recommended',
    'plugin:testing-library/dom',
    'plugin:testing-library/react',
  ],
  plugins: ['lodash', 'prefer-arrow', 'testing-library', 'jest'],
  env: {
    browser: true,
    jasmine: true,
    jest: true,
    es6: true,
  },
  globals: {
    fetchMock: true,
    process: true,
  },
  overrides: [],
  rules: {
    'import/default': 'off',
    'import/named': 'off',
    'import/no-anonymous-default-export': 'off',
    'import/order': [
      'error',
      {
        groups: [
          'builtin',
          'external',
          'internal',
          'parent',
          'sibling',
          'index',
        ],
        pathGroupsExcludedImportTypes: [],
        'newlines-between': 'always',
        alphabetize: {
          order: 'asc',
          caseInsensitive: true,
        },
      },
    ],
    'lodash/import-scope': ['error', 'method'],
    'react/prop-types': 'off',
    'react/self-closing-comp': [
      'error',
      {
        component: true,
        html: true,
      },
    ],
    '@typescript-eslint/no-unused-vars': [
      'error',
      {
        argsIgnorePattern: '^_',
        varsIgnorePattern: '^_',
        caughtErrorsIgnorePattern: '^_',
      },
    ],
    '@typescript-eslint/explicit-function-return-type': 'off',
    '@typescript-eslint/explicit-module-boundary-types': 'off',
    '@typescript-eslint/naming-convention': 'off',
    'react-hooks/exhaustive-deps': 'error',
    'testing-library/consistent-data-testid': [
      'error',
      {
        testIdPattern: '^{fileName}__[a-z]+\\w+$',
      },
    ],
    'no-useless-return': 'off',
    'jest/prefer-to-be': 'warn',
    'testing-library/no-node-access': 'off',
    'testing-library/no-unnecessary-act': 'off',
    'jest/expect-expect': 'off',
    'jest-dom/prefer-in-document': 'off',
  },
  settings: {
    react: {
      pragma: 'React',
      version: 'detect',
    },
    'import/extensions': ['.js', '.jsx', '.ts', '.tsx', '.css'],
    'import/parsers': {
      '@typescript-eslint/parser': ['.ts', '.tsx'],
    },
    'import/resolver': {
      typescript: {
        alwaysTryTypes: true,
      },
    },
  },
  parser: '@typescript-eslint/parser',
};
