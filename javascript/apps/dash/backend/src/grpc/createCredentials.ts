import {credentials} from '@grpc/grpc-js';

const createCredentials = (address: string) => {
  if (address.includes('localhost')) {
    return credentials.createInsecure();
  }

  return credentials.createSsl();
};

export default createCredentials;
