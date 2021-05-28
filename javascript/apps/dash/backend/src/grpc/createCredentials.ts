import {credentials} from '@grpc/grpc-js';

const createCredentials = () => {
  const {GRPC_SSL} = process.env;

  if (GRPC_SSL === 'false') {
    return credentials.createInsecure();
  }

  return credentials.createSsl();
};

export default createCredentials;
