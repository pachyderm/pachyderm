import {Repo, RepoInfo} from '@pachyderm/proto/pb/client/pfs/pfs_pb';

const repos: {[projectId: string]: RepoInfo[]} = {
  tutorial: [
    new RepoInfo().setRepo(new Repo().setName('montage')),
    new RepoInfo().setRepo(new Repo().setName('edges')),
    new RepoInfo().setRepo(new Repo().setName('images')),
  ],
};

export default repos;
