import {APIClient} from '@pachyderm/proto/pb/client/pfs/pfs_grpc_pb';
import {ListRepoRequest, RepoInfo} from '@pachyderm/proto/pb/client/pfs/pfs_pb';

import createCredentials from './createCredentials';

const pfs = (address: string, authToken: string) => {
  const client = new APIClient(address, createCredentials(authToken));

  return {
    listRepo: () => {
      return new Promise<RepoInfo.AsObject[]>((resolve, reject) => {
        client.listRepo(new ListRepoRequest(), (err, res) => {
          if (err) return reject(err);

          return resolve(res.toObject().repoInfoList);
        });
      });
    },
  };
};

export default pfs;
