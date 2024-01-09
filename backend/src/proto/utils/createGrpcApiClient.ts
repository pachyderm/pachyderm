import {credentials} from '@grpc/grpc-js';

const createCredentials = (ssl: boolean) => {
  if (ssl === false) {
    return credentials.createInsecure();
  }

  return credentials.createSsl();
};

export const grpcApiConstructorArgs = () => {
  const pachdAddress = process.env.PACHD_ADDRESS;
  if (!pachdAddress)
    throw new Error('env var PACHD_ADDRESS is not set. Set it!');

  const ssl = process.env.GRPC_SSL === 'true';
  const channelCredentials = createCredentials(ssl);

  return [pachdAddress, channelCredentials] as const;
};
