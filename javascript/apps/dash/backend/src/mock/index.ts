import {Server, ServerCredentials} from '@grpc/grpc-js';
import {APIService as PFSService} from '@pachyderm/proto/pb/client/pfs/pfs_grpc_pb';
import {APIService as PPSService} from '@pachyderm/proto/pb/client/pps/pps_grpc_pb';

import pfs from './handlers/pfs';
import pps from './handlers/pps';

const port = 50051;

const createServer = () => {
  const server = new Server();

  server.addService(PPSService, pps);
  server.addService(PFSService, pfs);

  const mockServer = {
    port,
    start: () => {
      return new Promise<number>((res, rej) => {
        server.bindAsync(
          `localhost:${port}`,
          ServerCredentials.createInsecure(),
          (err: Error | null, port: number) => {
            if (err != null) {
              rej(err);
            }

            console.log(`mock grpc listening on ${port}`);
            server.start();

            res(port);
          },
        );
      });
    },

    stop: () => {
      return new Promise<null>((res, rej) => {
        server.tryShutdown((err) => {
          if (err != null) {
            rej(err);
          }

          res(null);
        });
      });
    },
  };

  return mockServer;
};

const server = createServer();

if (require.main === module) {
  server.start();
}

export default server;
