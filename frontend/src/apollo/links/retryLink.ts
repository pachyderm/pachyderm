import {ServerError} from '@apollo/client';
import {RetryLink} from '@apollo/client/link/retry';
import {OperationDefinitionNode} from 'graphql';

import {
  TEST_MAX_RETRY_ATTEMPTS,
  TEST_MAX_RETRY_DELAY_MS,
} from '../../constants/retry';

export const retryLink = () =>
  new RetryLink({
    delay: {
      jitter: true,
      initial: 300,
      max: process.env.NODE_ENV === 'test' ? TEST_MAX_RETRY_DELAY_MS : Infinity,
    },
    attempts: {
      max: process.env.NODE_ENV === 'test' ? TEST_MAX_RETRY_ATTEMPTS : 5,
      retryIf: (error, operation) => {
        const {operation: operationType} = operation.query
          .definitions[0] as OperationDefinitionNode;
        if (error && operationType === 'query') {
          const {statusCode} = error as ServerError;
          return statusCode >= 500;
        }
        if (error && operationType === 'subscription') {
          return true;
        }
        return false;
      },
    },
  });
