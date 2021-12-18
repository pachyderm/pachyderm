/* eslint-disable @typescript-eslint/no-var-requires */
const path = require('path');

const {GraphQLFileLoader} = require('@graphql-tools/graphql-file-loader');
const {loadSchemaSync} = require('@graphql-tools/load');
const {printSchema} = require('graphql');

const schemaString = printSchema(
  loadSchemaSync(path.join(__dirname, '../backend/src/schema.graphqls'), {
    loaders: [new GraphQLFileLoader()],
  }),
);

module.exports = {
  extends: require.resolve('@pachyderm/config/eslint.config'),
  rules: {
    'graphql/template-strings': ['error', {env: 'apollo', schemaString}],
  },
  plugins: ['graphql'],
};
