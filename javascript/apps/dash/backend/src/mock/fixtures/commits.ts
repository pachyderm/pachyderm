import {Branch, Commit, CommitInfo} from '@pachyderm/proto/pb/pfs/pfs_pb';

import {timestampFromObject} from '@dash-backend/grpc/builders/protobuf';

import repos from './repos';

const tutorial = [
  new CommitInfo()
    .setCommit(
      new Commit()
        .setId('9d5daa0918ac4c43a476b86e3bb5e88e')
        .setBranch(
          new Branch().setName('master').setRepo(repos['3'][0].getRepo()),
        ),
    )
    .setDescription('added shinra hq building specs')
    .setSizeBytes(543211)
    .setStarted(timestampFromObject({seconds: 1614136189, nanos: 0}))
    .setFinished(timestampFromObject({seconds: 1614136191, nanos: 0})),
  new CommitInfo()
    .setCommit(
      new Commit()
        .setId('0918ac4c43a476b86e3bb5e88e9d5daa')
        .setBranch(
          new Branch().setName('master').setRepo(repos['3'][0].getRepo()),
        ),
    )
    .setDescription('added fire materia')
    .setSizeBytes(34371)
    .setStarted(timestampFromObject({seconds: 1614136289, nanos: 0}))
    .setFinished(timestampFromObject({seconds: 1614136291, nanos: 0})),
  new CommitInfo()
    .setCommit(
      new Commit()
        .setId('0918ac9d5daa76b86e3bb5e88e4c43a4')
        .setBranch(
          new Branch().setName('master').setRepo(repos['3'][0].getRepo()),
        ),
    )
    .setDescription('added mako')
    .setSizeBytes(44276)
    .setStarted(timestampFromObject({seconds: 1614136389, nanos: 0}))
    .setFinished(timestampFromObject({seconds: 1614136391, nanos: 0})),
];

const commits: {[projectId: string]: CommitInfo[]} = {
  '1': tutorial,
  '2': [],
  default: [...tutorial],
};

export default commits;
