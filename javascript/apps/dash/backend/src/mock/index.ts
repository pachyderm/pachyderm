import {Server, ServerCredentials} from '@grpc/grpc-js';
import {APIService as PFSService} from '@pachyderm/proto/pb/client/pfs/pfs_grpc_pb';
import {APIService as PPSService} from '@pachyderm/proto/pb/client/pps/pps_grpc_pb';

import pfs from './handlers/pfs';
import pps from './handlers/pps';

const port = 50051;

const startServer = () => {
  const server = new Server();

  server.addService(PPSService, pps);
  server.addService(PFSService, pfs);

  server.bindAsync(
    `0.0.0.0:${port}`,
    ServerCredentials.createInsecure(),
    (err: Error | null, port: number) => {
      if (err != null) {
        console.log(err);
      }

      console.log(`mock grpc listening on ${port}`);
      server.start();
    },
  );
};

startServer();
