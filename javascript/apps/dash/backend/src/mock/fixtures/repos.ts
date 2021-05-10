import {Branch, Repo, RepoInfo} from '@pachyderm/proto/pb/pfs/pfs_pb';

import {timestampFromObject} from '@dash-backend/grpc/builders/protobuf';

const tutorial = [
  new RepoInfo()
    .setRepo(new Repo().setName('montage'))
    .setSizeBytes(1000)
    .setCreated(timestampFromObject({seconds: 1614136189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('edges'))
    .setSizeBytes(1000)
    .setCreated(timestampFromObject({seconds: 1614126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('images'))
    .setSizeBytes(1000)
    .setCreated(timestampFromObject({seconds: 1614116189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
];

const customerTeam = [
  new RepoInfo()
    .setRepo(new Repo().setName('samples'))
    .setCreated(timestampFromObject({seconds: 1615326189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('likelihoods'))
    .setCreated(timestampFromObject({seconds: 1615526189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('reference'))
    .setCreated(timestampFromObject({seconds: 1615426189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('training'))
    .setCreated(timestampFromObject({seconds: 1615026189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('models'))
    .setCreated(timestampFromObject({seconds: 1615126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('joint_call'))
    .setCreated(timestampFromObject({seconds: 1615626189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('raw_data'))
    .setCreated(timestampFromObject({seconds: 1615626189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('split'))
    .setCreated(timestampFromObject({seconds: 1614726189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('parameters'))
    .setCreated(timestampFromObject({seconds: 1614626189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('model'))
    .setCreated(timestampFromObject({seconds: 1614526189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('test'))
    .setCreated(timestampFromObject({seconds: 1614426189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('select'))
    .setCreated(timestampFromObject({seconds: 1614326189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('detect'))
    .setCreated(timestampFromObject({seconds: 1614226189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('images'))
    .setCreated(timestampFromObject({seconds: 1614126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
];

const cron = [
  new RepoInfo()
    .setRepo(new Repo().setName('cron'))
    .setCreated(timestampFromObject({seconds: 1614126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('processor'))
    .setCreated(timestampFromObject({seconds: 1614226189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
];

const repos: {[projectId: string]: RepoInfo[]} = {
  '1': tutorial,
  '2': customerTeam,
  '3': cron,
  '4': customerTeam,
  '5': tutorial,
  default: [...tutorial, ...customerTeam],
};

export default repos;
