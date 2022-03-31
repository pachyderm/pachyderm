import {ApolloServerPlugin} from 'apollo-server-plugin-base';

import {Context} from '@dash-backend/lib/types';

const loggingPlugin: ApolloServerPlugin<Context> = {
  requestDidStart: ({request, context}) => {
    return new Promise((resolve) => {
      context.log.info(
        {
          eventSource: 'apollo-server',
          meta: {
            operationName: request.operationName,
          },
        },
        'request did start',
      );

      resolve({
        didEncounterErrors: ({request, context, errors}) => {
          return new Promise((resolve) => {
            context.log.error(
              {
                eventSource: 'apollo-server',
                meta: {
                  operationName: request.operationName,
                  errors,
                },
              },
              'did encounter errors',
            );
            resolve();
          });
        },
        willSendResponse: ({request, context}) => {
          return new Promise((resolve) => {
            context.log.info(
              {
                eventSource: 'apollo-server',
                meta: {
                  operationName: request.operationName,
                },
              },
              'will send response',
            );
            resolve();
          });
        },
      });
    });
  },
};

export default loggingPlugin;
