/* eslint-disable @typescript-eslint/no-var-requires */


module.exports = {
  extends: require.resolve('@pachyderm/config/eslint.config'),
  root: true,
  overrides: [
    {
      files: ['./src/queries/*'],
      processor: '@graphql-eslint/graphql',
      extends: ['eslint:recommended'],
      env: {
        node: true,
        es6: true,
      },
    },
    {
      files: ['*.graphqls'],
      parser: '@graphql-eslint/eslint-plugin',
      plugins: ['@graphql-eslint'],
      rules: {
        '@graphql-eslint/no-anonymous-operations': 'error',
      },
      parserOptions: {
        schema: '../backend/src/schema.graphqls',
      },
    },
  ],
  rules: {
    '@typescript-eslint/naming-convention': 'off',
  },
};
