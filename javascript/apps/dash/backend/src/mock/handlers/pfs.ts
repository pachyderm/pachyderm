import {Status} from '@grpc/grpc-js/build/src/constants';
import {IAPIServer} from '@pachyderm/proto/pb/pfs/pfs_grpc_pb';
import {ListRepoResponse} from '@pachyderm/proto/pb/pfs/pfs_pb';

import commits from '@dash-backend/mock/fixtures/commits';
import repos from '@dash-backend/mock/fixtures/repos';

const pfs: Pick<IAPIServer, 'listRepo' | 'inspectRepo' | 'listCommit'> = {
  listRepo: (call, callback) => {
    const [projectId] = call.metadata.get('project-id');

    const reply = new ListRepoResponse();

    reply.setRepoInfoList(projectId ? repos[projectId.toString()] : repos['1']);

    callback(null, reply);
  },
  inspectRepo: (call, callback) => {
    const [projectId] = call.metadata.get('project-id');
    const repoName = call.request.getRepo()?.getName();
    const repo = (
      projectId ? repos[projectId.toString()] : repos['tutorial']
    ).find((r) => r.getRepo()?.getName() === repoName);

    if (repo) {
      callback(null, repo);
    } else {
      callback({code: Status.NOT_FOUND, details: 'repo not found'});
    }
  },
  listCommit: (call) => {
    const [projectId] = call.metadata.get('project-id');
    const repoName = call.request.getRepo()?.getName();
    const allCommits = commits[projectId.toString()] || commits['1'];

    allCommits.forEach((commit) => {
      if (commit.getBranch()?.getRepo()?.getName() === repoName) {
        call.write(commit);
      }
    });

    call.end();
  },
};

export default pfs;
