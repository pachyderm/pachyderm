import fs from 'fs';

import {ApolloError} from 'apollo-server-errors';
import fetch from 'node-fetch';

import mockServer from './mock';

import graphqlServer from '.';

const operations = fs.readFileSync('../generated/operations.gql', 'utf-8');

const executeOperation = async <T>(
  operationName: string,
  variables: Record<string, unknown> = {},
) => {
  const response = await fetch(
    `http://localhost:${process.env.GRAPHQL_PORT}/graphql`,
    {
      method: 'POST',
      body: JSON.stringify({
        operationName,
        query: operations,
        variables,
      }),
      headers: {
        'pachd-address': `localhost:${process.env.GRPC_PORT}`,
        'Content-Type': 'application/json',
      },
    },
  );

  const json = await response.json();

  return {
    data: json.data as T | null,
    errors: json.errors as ApolloError[] | undefined,
  };
};

const createOperation = async <T>(
  query: string,
  variables: Record<string, unknown> = {},
) => {
  const response = await fetch(
    `http://localhost:${process.env.GRAPHQL_PORT}/graphql`,
    {
      method: 'POST',
      body: JSON.stringify({
        query,
        variables,
      }),
      headers: {
        'pachd-address': `localhost:${process.env.GRPC_PORT}`,
        'Content-Type': 'application/json',
      },
    },
  );

  const json = await response.json();

  return json.data as T;
};

export {mockServer, graphqlServer, executeOperation, createOperation};
