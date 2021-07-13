import {credentials} from '@grpc/grpc-js';

const createCredentials = (ssl: boolean) => {
  if (ssl === false) {
    return credentials.createInsecure();
  }

  return credentials.createSsl();
};

export default createCredentials;
