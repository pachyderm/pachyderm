import {credentials} from '@grpc/grpc-js';

const createCredentials = (address: string) => {
  if (
    address.includes('localhost') ||
    address.includes('host.docker.internal')
  ) {
    return credentials.createInsecure();
  }

  return credentials.createSsl();
};

export default createCredentials;
