import {ChannelCredentials, Metadata} from '@grpc/grpc-js';
import {APIClient} from '@pachyderm/proto/pb/pfs/pfs_grpc_pb';
import {ListRepoRequest, RepoInfo} from '@pachyderm/proto/pb/pfs/pfs_pb';

const pfs = (
  pachdAddress: string,
  channelCredentials: ChannelCredentials,
  credentialMetadata: Metadata,
) => {
  const client = new APIClient(pachdAddress, channelCredentials);

  return {
    listRepo: () => {
      return new Promise<RepoInfo.AsObject[]>((resolve, reject) => {
        client.listRepo(
          new ListRepoRequest(),
          credentialMetadata,
          (err, res) => {
            if (err) return reject(err);

            return resolve(res.toObject().repoInfoList);
          },
        );
      });
    },
  };
};

export default pfs;
