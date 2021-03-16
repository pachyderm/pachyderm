import {ApolloServerPlugin} from 'apollo-server-plugin-base';

import {Context} from '@dash-backend/lib/types';

const loggingPlugin: ApolloServerPlugin<Context> = {
  requestDidStart: ({request, context}) => {
    context.log.info(
      {
        eventSource: 'apollo-server',
        meta: {
          operationName: request.operationName,
        },
      },
      'request did start',
    );

    return {
      didEncounterErrors: ({request, context, errors}) => {
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
      },
      willSendResponse: ({request, context}) => {
        context.log.info(
          {
            eventSource: 'apollo-server',
            meta: {
              operationName: request.operationName,
            },
          },
          'will send response',
        );
      },
    };
  },
};

export default loggingPlugin;
