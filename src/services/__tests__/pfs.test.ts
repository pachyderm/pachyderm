import client from '../../client';
import {CommitState, FileType} from '../../proto/pfs/pfs_pb';

describe('services/pfs', () => {
  afterAll(async () => {
    const pachClient = client({ssl: false, pachdAddress: 'localhost:30650'});
    const pfs = pachClient.pfs();
    await pfs.deleteAll();
  });
  const createSandbox = async (name: string) => {
    const pachClient = client({ssl: false, pachdAddress: 'localhost:30650'});
    const pfs = pachClient.pfs();
    await pfs.deleteAll();
    await pfs.createRepo({repo: {name}});

    return pachClient;
  };
  describe('listFile', () => {
    it('should return a list of files in the directory', async () => {
      const client = await createSandbox('listFile');
      const commit = await client.pfs().startCommit({
        branch: {name: 'master', repo: {name: 'listFile'}},
      });

      await client
        .modifyFile()
        .setCommit(commit)
        .putFileFromURL('at-at.png', 'http://imgur.com/8MN9Kg0.png')
        .end();

      await client.pfs().finishCommit({commit});

      const files = await client.pfs().listFile({
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'listFile'}},
      });

      expect(files).toHaveLength(1);
    });
  });

  describe('getFile', () => {
    it('should return a file from a repo', async () => {
      const client = await createSandbox('getFile');
      const commit = await client.pfs().startCommit({
        branch: {name: 'master', repo: {name: 'getFile'}},
      });

      await client
        .modifyFile()
        .setCommit(commit)
        .putFileFromURL('at-at.png', 'http://imgur.com/8MN9Kg0.png')
        .end();

      await client.pfs().finishCommit({commit});
      const file = await client.pfs().getFile({
        commitId: commit.id,
        path: '/at-at.png',
        branch: {name: 'master', repo: {name: 'getFile'}},
      });
      expect(file.byteLength).toEqual(80588);
    });

    it('should return a file TAR from a repo', async () => {
      const client = await createSandbox('getFile');
      const commit = await client.pfs().startCommit({
        branch: {name: 'master', repo: {name: 'getFile'}},
      });

      await client
        .modifyFile()
        .setCommit(commit)
        .putFileFromURL('at-at.png', 'http://imgur.com/8MN9Kg0.png')
        .end();

      await client.pfs().finishCommit({commit});
      const file = await client.pfs().getFileTAR({
        commitId: commit.id,
        path: '/at-at.png',
        branch: {name: 'master', repo: {name: 'getFile'}},
      });
      expect(file.byteLength).toEqual(82432);
    });
  });

  describe('inspectFile', () => {
    it('should return details about the specified file', async () => {
      const client = await createSandbox('inspectRepo');
      const commit = await client.pfs().startCommit({
        branch: {name: 'master', repo: {name: 'inspectRepo'}},
      });

      await client
        .modifyFile()
        .setCommit(commit)
        .putFileFromURL('at-at.png', 'http://imgur.com/8MN9Kg0.png')
        .end();

      await client.pfs().finishCommit({commit});
      const file = await client.pfs().inspectFile({
        commitId: commit.id,
        path: '/at-at.png',
        branch: {name: 'master', repo: {name: 'inspectRepo'}},
      });

      expect(file.file?.commit?.branch?.name).toEqual('master');
      expect(file.file?.commit?.id).toEqual(commit.id);
      expect(file.fileType).toEqual(FileType.FILE);
      expect(file.sizeBytes).toEqual(80588);
    });
  });
  describe('listCommit', () => {
    it('should list the commits for the specified repo', async () => {
      const client = await createSandbox('listCommit');
      const commit = await client
        .pfs()
        .startCommit({branch: {name: 'master', repo: {name: 'listCommit'}}});
      await client.pfs().finishCommit({commit});
      await client
        .pfs()
        .startCommit({branch: {name: 'master', repo: {name: 'listCommit'}}});

      const commits = await client
        .pfs()
        .listCommit({repo: {name: 'listCommit'}});

      expect(commits).toHaveLength(2);
    });
    it('should return only the specified number of commits if number is specified', async () => {
      const client = await createSandbox('listCommit');
      const commit = await client
        .pfs()
        .startCommit({branch: {name: 'master', repo: {name: 'listCommit'}}});
      await client.pfs().finishCommit({commit});
      await client
        .pfs()
        .startCommit({branch: {name: 'master', repo: {name: 'listCommit'}}});

      const commits = await client
        .pfs()
        .listCommit({repo: {name: 'listCommit'}, number: 1});

      expect(commits).toHaveLength(1);
    });
  });
  describe('inspectCommitSet', () => {
    it('should return details about a commit set', async () => {
      const client = await createSandbox('inspectCommitSet');
      const commit = await client.pfs().startCommit({
        branch: {name: 'master', repo: {name: 'inspectCommitSet'}},
      });
      await client.pfs().finishCommit({commit});
      const commitSet = await client
        .pfs()
        .inspectCommitSet({commitSet: commit});

      expect(commitSet).toHaveLength(1);
      expect(commitSet[0].error).toBeFalsy();
      expect(commitSet[0].details?.sizeBytes).toEqual(0);
    });
  });
  describe('listCommitSet', () => {
    it('should list the commit sets', async () => {
      const client = await createSandbox('listCommitSet');
      const commit = await client.pfs().startCommit({
        branch: {name: 'master', repo: {name: 'listCommitSet'}},
      });
      await client.pfs().finishCommit({commit});

      const commitSets = await client.pfs().listCommitSet();
      expect(commitSets).toHaveLength(1);
    });
  });
  describe('squashCommitSet', () => {
    it('should squash commits', async () => {
      const client = await createSandbox('squashCommitSet');
      const commit1 = await client.pfs().startCommit({
        branch: {name: 'master', repo: {name: 'squashCommitSet'}},
      });
      await client.pfs().finishCommit({commit: commit1});
      const commit2 = await client.pfs().startCommit({
        branch: {name: 'master', repo: {name: 'squashCommitSet'}},
      });
      await client.pfs().finishCommit({commit: commit2});
      await client
        .pfs()
        .inspectCommit({commit: commit1, wait: CommitState.FINISHED});
      await client
        .pfs()
        .inspectCommit({commit: commit2, wait: CommitState.FINISHED});

      const commits = await client
        .pfs()
        .listCommit({repo: {name: 'squashCommitSet'}});

      expect(commits).toHaveLength(2);

      await client.pfs().squashCommitSet(commit1);

      const updatedCommits = await client
        .pfs()
        .listCommit({repo: {name: 'squashCommitSet'}});

      expect(updatedCommits).toHaveLength(1);
    });
  });
  describe('createBranch', () => {
    it('should create a branch', async () => {
      const client = await createSandbox('createBranch');

      const commit = await client
        .pfs()
        .startCommit({branch: {name: 'master', repo: {name: 'createBranch'}}});

      await client.pfs().createBranch({
        newCommitSet: false,
        head: commit,
        provenance: [],
        branch: {name: 'test', repo: {name: 'createBranch'}},
      });
      const updatedBranches = await client
        .pfs()
        .listBranch({repo: {name: 'createBranch'}});

      expect(updatedBranches).toHaveLength(2);
      expect(updatedBranches[0].branch?.name).toEqual('test');
    });
  });
  describe('inspectBranch', () => {
    it('should return details about a branch', async () => {
      const client = await createSandbox('inspectBranch');
      await client
        .pfs()
        .startCommit({branch: {name: 'master', repo: {name: 'inspectBranch'}}});
      const branch = await client
        .pfs()
        .inspectBranch({name: 'master', repo: {name: 'inspectBranch'}});

      expect(branch.branch?.name).toEqual('master');
      expect(branch.head?.branch?.name).toEqual('master');
      expect(branch.provenanceList).toHaveLength(0);
      expect(branch.subvenanceList).toHaveLength(0);
      expect(branch.directProvenanceList).toHaveLength(0);
    });
  });
  describe('listBranch', () => {
    it('should return a list of branches', async () => {
      const client = await createSandbox('listBranch');
      await client
        .pfs()
        .startCommit({branch: {name: 'master', repo: {name: 'listBranch'}}});
      const branches = await client
        .pfs()
        .listBranch({repo: {name: 'listBranch'}});
      expect(branches).toHaveLength(1);

      expect(branches).toHaveLength(1);
      expect(branches[0].branch?.name).toEqual('master');
      expect(branches[0].provenanceList).toHaveLength(0);
      expect(branches[0].subvenanceList).toHaveLength(0);
    });
  });
  describe('deleteBranch', () => {
    it('should delete branch', async () => {
      const client = await createSandbox('deleteBranch');
      const initialBranches = await client
        .pfs()
        .listBranch({repo: {name: 'deleteBranch'}});
      expect(initialBranches).toHaveLength(0);

      await client
        .pfs()
        .startCommit({branch: {name: 'master', repo: {name: 'deleteBranch'}}});

      const updatedBranches = await client
        .pfs()
        .listBranch({repo: {name: 'deleteBranch'}});

      expect(updatedBranches).toHaveLength(1);

      await client
        .pfs()
        .deleteBranch({branch: {name: 'master', repo: {name: 'deleteBranch'}}});

      const finalBranches = await client
        .pfs()
        .listBranch({repo: {name: 'deleteBranch'}});

      expect(finalBranches).toHaveLength(0);
    });
  });
  describe('listRepo', () => {
    it('should return a list of all repos in the cluster', async () => {
      const client = await createSandbox('listRepo');
      const repos = await client.pfs().listRepo();

      expect(repos).toHaveLength(1);
    });
  });
  describe('inspectRepo', () => {
    it('should return information about the specified', async () => {
      const client = await createSandbox('inspectRepo');
      const repo = await client.pfs().inspectRepo('inspectRepo');

      expect(repo.repo?.name).toEqual('inspectRepo');
      expect(repo.repo?.type).toEqual('user');
      expect(repo.branchesList).toHaveLength(0);
    });
  });
  describe('createRepo', () => {
    it('should create a repo', async () => {
      const client = await createSandbox('createRepo');

      const initialRepos = await client.pfs().listRepo();
      expect(initialRepos).toHaveLength(1);

      await client.pfs().createRepo({repo: {name: 'anotherRepo'}});

      const udatedRepos = await client.pfs().listRepo();
      expect(udatedRepos).toHaveLength(2);
    });
  });
  describe('deleteRepo', () => {
    it('should delete a repo', async () => {
      const client = await createSandbox('deleteRepo');

      const initialRepos = await client.pfs().listRepo();
      expect(initialRepos).toHaveLength(1);

      await client.pfs().deleteRepo({repo: {name: 'deleteRepo'}});

      const updatedRepos = await client.pfs().listRepo();
      expect(updatedRepos).toHaveLength(0);
    });
  });
});
