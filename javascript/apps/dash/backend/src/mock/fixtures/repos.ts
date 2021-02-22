import {Repo, RepoInfo} from '@pachyderm/proto/pb/pfs/pfs_pb';

const repos: {[projectId: string]: RepoInfo[]} = {
  tutorial: [
    new RepoInfo().setRepo(new Repo().setName('montage')),
    new RepoInfo().setRepo(new Repo().setName('edges')),
    new RepoInfo().setRepo(new Repo().setName('images')),
  ],
  customerTeam: [
    new RepoInfo().setRepo(new Repo().setName('samples')),
    new RepoInfo().setRepo(new Repo().setName('likelihoods')),
    new RepoInfo().setRepo(new Repo().setName('reference')),
    new RepoInfo().setRepo(new Repo().setName('training')),
    new RepoInfo().setRepo(new Repo().setName('models')),
    new RepoInfo().setRepo(new Repo().setName('joint_call')),
    new RepoInfo().setRepo(new Repo().setName('raw_data')),
    new RepoInfo().setRepo(new Repo().setName('split')),
    new RepoInfo().setRepo(new Repo().setName('parameters')),
    new RepoInfo().setRepo(new Repo().setName('model')),
    new RepoInfo().setRepo(new Repo().setName('test')),
    new RepoInfo().setRepo(new Repo().setName('select')),
    new RepoInfo().setRepo(new Repo().setName('detect')),
    new RepoInfo().setRepo(new Repo().setName('images')),
  ],
};

export default repos;
