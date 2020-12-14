import {credentials, Metadata} from '@grpc/grpc-js';

const createCredentials = (authToken: string) => {
  const channelCredentials = credentials.createSsl();
  const callCredentials = credentials.createFromMetadataGenerator(
    (_, callback) => {
      const meta = new Metadata();

      meta.add('authn-token', authToken);
      callback(null, meta);
    },
  );

  return credentials.combineChannelCredentials(
    channelCredentials,
    callCredentials,
  );
};

export default createCredentials;
