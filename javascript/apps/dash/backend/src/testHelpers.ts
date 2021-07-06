import fs from 'fs';
import path from 'path';

import {
  ApolloClient,
  ApolloQueryResult,
  DocumentNode,
  FetchResult,
  HttpLink,
  InMemoryCache,
  Observable,
} from '@apollo/client/core';
import {Metadata, StatusBuilder, status} from '@grpc/grpc-js';
import {callErrorFromStatus} from '@grpc/grpc-js/build/src/call';
import {ApolloError} from 'apollo-server-errors';
import fetch from 'cross-fetch';
import {print} from 'graphql';
import {SubscribePayload, createClient} from 'graphql-ws';
import {sign} from 'jsonwebtoken';
import {v4 as uuid} from 'uuid';
import ws from 'ws';

import mockServer from '@dash-backend/mock';
import keys from '@dash-backend/mock/fixtures/keys';
import {Account} from '@graphqlTypes';

import graphqlServer from '.';

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

const createSubscriptionClients = <T>(
  query: DocumentNode,
  variables: SubscribePayload['variables'],
) => {
  const client = createClient({
    url: `ws://localhost:${process.env.GRAPHQL_PORT}/graphql`,
    connectionParams: () => {
      return {
        'id-token': generateIdTokenForAccount(mockServer.getAccount()),
        'pachd-address': `localhost:${process.env.GRPC_PORT}`,
        'auth-token': mockServer.getAccount().id,
      };
    },
    webSocketImpl: ws,
    generateID: uuid,
    keepAlive: 0,
    retryAttempts: 0,
  });

  const toObservable = () => {
    return new Observable<T>((observer) =>
      client.subscribe<T>(
        {query: print(query), variables},
        {
          next: (data) => observer.next(data),
          error: (err) => observer.error(err),
          complete: () => observer.complete(),
        },
      ),
    );
  };

  const observable = toObservable();

  return {observable};
};

const executeQuery = async <T>(
  query: DocumentNode,
  variables: Record<string, unknown> = {},
  headers: {[key: string]: string} = {},
) => {
  const link = new HttpLink({
    uri: `http://localhost:${process.env.GRAPHQL_PORT}/graphql`,
    fetch: fetch,
  });

  const client = new ApolloClient({cache: new InMemoryCache(), link});

  const context = {
    headers: {
      'id-token': generateIdTokenForAccount(mockServer.getAccount()),
      'auth-token': mockServer.getAccount().id,
      'Content-Type': 'application/json',
      ...headers,
    },
  };

  let data: ApolloQueryResult<T> | undefined;
  let errors: ApolloError[] | undefined;
  try {
    data = await client.query<T>({query, variables, context});
  } catch (err) {
    errors = err.graphQLErrors;
  }

  return {data: data?.data || null, errors};
};

const executeMutation = async <T>(
  mutation: DocumentNode,
  variables: Record<string, unknown> = {},
  headers: {[key: string]: string} = {},
) => {
  const link = new HttpLink({
    uri: `http://localhost:${process.env.GRAPHQL_PORT}/graphql`,
    fetch: fetch,
  });

  const client = new ApolloClient({cache: new InMemoryCache(), link});

  const context = {
    headers: {
      'id-token': generateIdTokenForAccount(mockServer.getAccount()),
      'auth-token': mockServer.getAccount().id,
      'Content-Type': 'application/json',
      ...headers,
    },
  };

  let data:
    | FetchResult<T, Record<string, unknown>, Record<string, unknown>>
    | undefined;
  let errors: ApolloError[] | undefined;
  try {
    data = await client.mutate<T>({mutation, variables, context});
  } catch (err) {
    errors = err.graphQLErrors;
  }

  return {data: data?.data || null, errors};
};

export {
  createSubscriptionClients,
  executeQuery,
  executeMutation,
  status,
  createServiceError,
  mockServer,
  graphqlServer,
};
