import fs from 'fs';
import path from 'path';

import {Metadata, StatusBuilder, status} from '@grpc/grpc-js';
import {callErrorFromStatus} from '@grpc/grpc-js/build/src/call';
import {ApolloError} from 'apollo-server-errors';
import {sign} from 'jsonwebtoken';
import fetch from 'node-fetch';

import mockServer from '@dash-backend/mock';
import keys from '@dash-backend/mock/fixtures/keys';
import {Account} from '@graphqlTypes';

import graphqlServer from '.';

const operations = fs.readFileSync('../generated/operations.gql', 'utf-8');

const executeOperation = async <T>(
  operationName: string,
  variables: Record<string, unknown> = {},
  headers: {[key: string]: string} = {},
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
        'id-token': generateIdTokenForAccount(mockServer.state.account),
        'Content-Type': 'application/json',
        ...headers,
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
  headers: {[key: string]: string} = {},
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
        'id-token': generateIdTokenForAccount(mockServer.state.account),
        'Content-Type': 'application/json',
        ...headers,
      },
    },
  );

  const json = await response.json();

  return {
    data: json.data as T | null,
    errors: json.errors as ApolloError[] | undefined,
  };
};

export const generateIdTokenForAccount = ({id, ...rest}: Account) => {
  return sign(
    {some: 'stuff', azp: 'dash', ...rest},
    fs.readFileSync(path.resolve(__dirname, 'mock/mockPrivate.key')),
    {
      algorithm: 'RS256',
      issuer: process.env.ISSUER_URI,
      subject: id,
      audience: ['pachd', 'dash'],
      expiresIn: '30 days',
      keyid: keys.keys[0].kid,
    },
  );
};

const createServiceError = (statusArgs: {
  code: status;
  details?: string;
  metadata?: Metadata;
}) => {
  const defaultStatusArgs = {
    code: status.INTERNAL,
    details: 'error',
    metadata: new Metadata(),
  };

  const statusObj = new StatusBuilder()
    .withCode(statusArgs.code)
    .withDetails(statusArgs.details || defaultStatusArgs.details)
    .withMetadata(statusArgs.metadata || defaultStatusArgs.metadata)
    .build();

  // callErrorFromStatus expects a StatusObject but StatusBuilder returns Partial<StatusObject>
  return callErrorFromStatus({...defaultStatusArgs, ...statusObj});
};

export {
  status,
  createServiceError,
  mockServer,
  graphqlServer,
  executeOperation,
  createOperation,
};
