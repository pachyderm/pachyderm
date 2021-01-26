import {Repo, RepoInfo} from '@pachyderm/proto/pb/client/pfs/pfs_pb';

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
  ],
};

export default repos;
