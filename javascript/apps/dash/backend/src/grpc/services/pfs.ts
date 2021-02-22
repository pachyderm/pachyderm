import {ChannelCredentials, Metadata} from '@grpc/grpc-js';
import {APIClient} from '@pachyderm/proto/pb/pfs/pfs_grpc_pb';
import {ListRepoRequest, RepoInfo} from '@pachyderm/proto/pb/pfs/pfs_pb';

const pfs = (pachdAddress: string, channelCredentials: ChannelCredentials) => {
  const client = new APIClient(pachdAddress, channelCredentials);

  return {
    listRepo: (projectId = '') => {
      return new Promise<RepoInfo.AsObject[]>((resolve, reject) => {
        const metadata = new Metadata();
        metadata.set('project-id', projectId);

        client.listRepo(new ListRepoRequest(), metadata, (err, res) => {
          if (err) return reject(err);

          return resolve(res.toObject().repoInfoList);
        });
      });
    },
  };
};

export default pfs;
