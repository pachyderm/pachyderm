import {IAPIServer} from '@pachyderm/proto/pb/client/pfs/pfs_grpc_pb';
import {ListRepoResponse} from '@pachyderm/proto/pb/client/pfs/pfs_pb';

import repos from 'mock/fixtures/repos';

const pfs: Pick<IAPIServer, 'listRepo'> = {
  listRepo: (call, callback) => {
    const [projectId] = call.metadata.get('project-id');

    const reply = new ListRepoResponse();

    // "tutorial" in this case represents the default/catch-all project in core pach
    reply.setRepoInfoList(
      projectId ? repos[projectId.toString()] : repos['tutorial'],
    );

    callback(null, reply);
  },
};

export default pfs;
