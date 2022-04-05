import {ApolloServerPlugin} from 'apollo-server-plugin-base';

const httpPlugin: ApolloServerPlugin = {
  async requestDidStart() {
    return {
      async willSendResponse({response}) {
        if (response && response.http) {
          if (
            response.errors?.some(
              (error) =>
                error.extensions?.code &&
                error.extensions.code === 'INTERNAL_SERVER_ERROR',
            )
          )
            response.http.status = 500;
        }
      },
    };
  },
};

export default httpPlugin;
