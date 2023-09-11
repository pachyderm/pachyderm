import path from 'path';

import apiClientRequestWrapper from '../../../client';
import {CommitState, FileType} from '../../../proto/pfs/pfs_pb';

const libertyPngFilePath = path.resolve(
  __dirname,
  '../../../../../../etc/testing/files/liberty.png',
);
const atatPngFilePath = path.resolve(
  __dirname,
  '../../../../../../etc/testing/files/AT-AT.png',
);

jest.setTimeout(30_000);

describe('services/pfs', () => {
  afterAll(async () => {
    const pachClient = apiClientRequestWrapper();
    const pfs = pachClient.pfs;
    await pfs.deleteAll();
  });
  const createSandbox = async (name: string) => {
    const pachClient = apiClientRequestWrapper();
    const pfs = pachClient.pfs;
    await pfs.deleteAll();
    await pfs.createRepo({projectId: 'default', repo: {name}});

    return pachClient;
  };
  describe('listFile', () => {
    it('should return a list of files in the directory', async () => {
      const client = await createSandbox('listFile');
      const commit = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'listFile'}},
      });

      const fileClient = await client.pfs.modifyFile();

      await fileClient
        .setCommit(commit)
        .putFileFromFilepath(atatPngFilePath, 'at-at.png')
        .end();

      await client.pfs.finishCommit({projectId: 'default', commit});

      const files = await client.pfs.listFile({
        projectId: 'default',
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'listFile'}},
      });

      expect(files).toHaveLength(1);
    });
  });

  describe('getFile', () => {
    it('should return a file from a repo', async () => {
      const client = await createSandbox('getFile');
      const commit = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'getFile'}},
      });
      const fileClient = await client.pfs.modifyFile();

      await fileClient
        .setCommit(commit)
        .putFileFromFilepath(atatPngFilePath, 'at-at.png')
        .end();

      await client.pfs.finishCommit({projectId: 'default', commit});
      const file = await client.pfs.getFile({
        projectId: 'default',
        commitId: commit.id,
        path: '/at-at.png',
        branch: {name: 'master', repo: {name: 'getFile'}},
      });
      expect(file.byteLength).toBe(80588);
    });

    it('should return a file TAR from a repo', async () => {
      const client = await createSandbox('getFile');
      const commit = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'getFile'}},
      });
      const fileClient = await client.pfs.modifyFile();

      await fileClient
        .setCommit(commit)
        .putFileFromFilepath(atatPngFilePath, 'at-at.png')
        .end();

      await client.pfs.finishCommit({projectId: 'default', commit});
      const file = await client.pfs.getFileTAR({
        projectId: 'default',
        commitId: commit.id,
        path: '/at-at.png',
        branch: {name: 'master', repo: {name: 'getFile'}},
      });
      expect(file.byteLength).toBe(82432);
    });
  });

  describe('inspectFile', () => {
    it('should return details about the specified file', async () => {
      const client = await createSandbox('inspectRepo');
      const commit = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'inspectRepo'}},
      });
      const fileClient = await client.pfs.modifyFile();

      await fileClient
        .setCommit(commit)
        .putFileFromFilepath(atatPngFilePath, 'at-at.png')
        .end();

      await client.pfs.finishCommit({projectId: 'default', commit});
      const file = await client.pfs.inspectFile({
        projectId: 'default',
        commitId: commit.id,
        path: '/at-at.png',
        branch: {name: 'master', repo: {name: 'inspectRepo'}},
      });

      expect(file.file?.commit?.branch?.name).toBe('master');
      expect(file.file?.commit?.id).toEqual(commit.id);
      expect(file.fileType).toEqual(FileType.FILE);
      expect(file.sizeBytes).toBe(80588);
    });
  });
  describe('listCommit', () => {
    it('should list the commits for the specified repo', async () => {
      const client = await createSandbox('listCommit');
      const commit = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'listCommit'}},
      });
      await client.pfs.finishCommit({projectId: 'default', commit});
      await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'listCommit'}},
      });

      const commits = await client.pfs.listCommit({
        projectId: 'default',
        repo: {name: 'listCommit'},
      });

      expect(commits).toHaveLength(2);
    });
    it('should return only the specified number of commits if number is specified', async () => {
      const client = await createSandbox('listCommit');
      const commit = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'listCommit'}},
      });
      await client.pfs.finishCommit({projectId: 'default', commit});
      await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'listCommit'}},
      });

      const commits = await client.pfs.listCommit({
        projectId: 'default',
        repo: {name: 'listCommit'},
        number: 1,
      });

      expect(commits).toHaveLength(1);
    });
  });

  describe('findCommits', () => {
    it('should return the commit history for the specified path in the repo', async () => {
      const client = await createSandbox('findCommits');

      const fileClient1 = await client.pfs.modifyFile();
      const commit1 = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'findCommits'}},
      });
      await fileClient1
        .setCommit(commit1)
        .putFileFromFilepath(atatPngFilePath, 'at-at.png')
        .end();
      await client.pfs.finishCommit({projectId: 'default', commit: commit1});
      await client.pfs.inspectCommit({
        projectId: 'default',
        commit: commit1,
        wait: CommitState.FINISHED,
      });

      const fileClient2 = await client.pfs.modifyFile();
      const commit2 = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'findCommits'}},
      });
      await fileClient2
        .setCommit(commit2)
        .putFileFromFilepath(libertyPngFilePath, 'at-at.png')
        .end();
      await client.pfs.finishCommit({projectId: 'default', commit: commit2});
      await client.pfs.inspectCommit({
        projectId: 'default',
        commit: commit2,
        wait: CommitState.FINISHED,
      });

      const commits = await client.pfs.findCommits({
        commit: {
          id: '',
          branch: {
            name: 'master',
            repo: {name: 'findCommits', project: {name: 'default'}},
          },
        },
        path: '/at-at.png',
      });

      expect(commits).toHaveLength(3);
      expect(commits[0].foundCommit?.id).toEqual(commit2.id);
      expect(commits[1].foundCommit?.id).toEqual(commit1.id);
      expect(commits[2].lastSearchedCommit?.id).toEqual(commit1.id);
    });
  });

  describe('inspectCommitSet', () => {
    it('should return details about a commit set', async () => {
      const client = await createSandbox('inspectCommitSet');
      const commit = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'inspectCommitSet'}},
      });
      await client.pfs.finishCommit({projectId: 'default', commit});
      const commitSet = await client.pfs.inspectCommitSet({commitSet: commit});

      expect(commitSet).toHaveLength(1);
      expect(commitSet[0].error).toBeFalsy();
      expect(commitSet[0].details?.sizeBytes).toBe(0);
    });
  });
  describe('listCommitSet', () => {
    it('should list the commit sets', async () => {
      const client = await createSandbox('listCommitSet');
      const commit = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'listCommitSet'}},
      });
      await client.pfs.finishCommit({projectId: 'default', commit});

      const commitSets = await client.pfs.listCommitSet();
      expect(commitSets).toHaveLength(1);
    });
  });
  describe('squashCommitSet', () => {
    it('should squash commits', async () => {
      const client = await createSandbox('squashCommitSet');
      const commit1 = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'squashCommitSet'}},
      });
      await client.pfs.finishCommit({projectId: 'default', commit: commit1});
      const commit2 = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'squashCommitSet'}},
      });
      await client.pfs.finishCommit({projectId: 'default', commit: commit2});
      await client.pfs.inspectCommit({
        projectId: 'default',
        commit: commit1,
        wait: CommitState.FINISHED,
      });
      await client.pfs.inspectCommit({
        projectId: 'default',
        commit: commit2,
        wait: CommitState.FINISHED,
      });

      const commits = await client.pfs.listCommit({
        projectId: 'default',
        repo: {name: 'squashCommitSet'},
      });

      expect(commits).toHaveLength(2);

      await client.pfs.squashCommitSet(commit1);

      const updatedCommits = await client.pfs.listCommit({
        projectId: 'default',
        repo: {name: 'squashCommitSet'},
      });

      expect(updatedCommits).toHaveLength(1);
    });
  });
  describe('dropCommitSet', () => {
    it('should remove the commit set', async () => {
      const client = await createSandbox('dropCommitSet');
      const commit1 = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'dropCommitSet'}},
      });
      await client.pfs.finishCommit({projectId: 'default', commit: commit1});

      const commit2 = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'dropCommitSet'}},
      });

      const commits = await client.pfs.listCommit({
        projectId: 'default',
        repo: {name: 'dropCommitSet'},
      });

      expect(commits).toHaveLength(2);

      await client.pfs.dropCommitSet({id: commit2.id});

      const updatedCommits = await client.pfs.listCommit({
        projectId: 'default',
        repo: {name: 'dropCommitSet'},
      });

      expect(updatedCommits).toHaveLength(1);
    });
  });
  describe('createBranch', () => {
    it('should create a branch', async () => {
      const client = await createSandbox('createBranch');

      const commit = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'createBranch'}},
      });

      await client.pfs.createBranch({
        newCommitSet: false,
        head: commit,
        provenance: [],
        branch: {name: 'test', repo: {name: 'createBranch'}},
      });
      const updatedBranches = await client.pfs.listBranch({
        projectId: 'default',
        repoName: 'createBranch',
      });

      expect(updatedBranches).toHaveLength(2);
      expect(updatedBranches[0].branch?.name).toBe('test');
    });
  });
  describe('inspectBranch', () => {
    it('should return details about a branch', async () => {
      const client = await createSandbox('inspectBranch');
      await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'inspectBranch'}},
      });
      const branch = await client.pfs.inspectBranch({
        projectId: 'default',
        name: 'master',
        repo: {name: 'inspectBranch'},
      });

      expect(branch.branch?.name).toBe('master');
      expect(branch.head?.branch?.name).toBe('master');
      expect(branch.provenanceList).toHaveLength(0);
      expect(branch.subvenanceList).toHaveLength(0);
      expect(branch.directProvenanceList).toHaveLength(0);
    });
  });
  describe('listBranch', () => {
    it('should return a list of branches', async () => {
      const client = await createSandbox('listBranch');
      await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'listBranch'}},
      });
      const branches = await client.pfs.listBranch({
        projectId: 'default',
        repoName: 'listBranch',
      });
      expect(branches).toHaveLength(1);

      expect(branches).toHaveLength(1);
      expect(branches[0].branch?.name).toBe('master');
      expect(branches[0].provenanceList).toHaveLength(0);
      expect(branches[0].subvenanceList).toHaveLength(0);
    });
  });
  describe('deleteBranch', () => {
    it('should delete branch', async () => {
      const client = await createSandbox('deleteBranch');
      const initialBranches = await client.pfs.listBranch({
        projectId: 'default',
        repoName: 'deleteBranch',
      });
      expect(initialBranches).toHaveLength(0);

      await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'deleteBranch'}},
      });

      const updatedBranches = await client.pfs.listBranch({
        projectId: 'default',
        repoName: 'deleteBranch',
      });

      expect(updatedBranches).toHaveLength(1);

      await client.pfs.deleteBranch({
        branch: {name: 'master', repo: {name: 'deleteBranch'}},
      });

      const finalBranches = await client.pfs.listBranch({
        projectId: 'default',
        repoName: 'deleteBranch',
      });

      expect(finalBranches).toHaveLength(0);
    });
  });
  describe('listRepo', () => {
    it('should return a list of all repos in the cluster', async () => {
      const client = await createSandbox('listRepo');
      const repos = await client.pfs.listRepo({projectIds: []});

      expect(repos).toHaveLength(1);
    });
  });
  describe('inspectRepo', () => {
    it('should return information about the specified', async () => {
      const client = await createSandbox('inspectRepo');
      const repo = await client.pfs.inspectRepo({
        projectId: 'default',
        name: 'inspectRepo',
      });

      expect(repo.repo?.name).toBe('inspectRepo');
      expect(repo.repo?.type).toBe('user');
      expect(repo.branchesList).toHaveLength(0);
    });
  });
  describe('createRepo', () => {
    it('should create a repo', async () => {
      const client = await createSandbox('createRepo');

      const initialRepos = await client.pfs.listRepo({projectIds: []});
      expect(initialRepos).toHaveLength(1);

      await client.pfs.createRepo({
        projectId: 'default',
        repo: {name: 'anotherRepo'},
      });

      const udatedRepos = await client.pfs.listRepo({projectIds: []});
      expect(udatedRepos).toHaveLength(2);
    });
  });
  describe('deleteRepo', () => {
    it('should delete a repo', async () => {
      const client = await createSandbox('deleteRepo');

      const initialRepos = await client.pfs.listRepo({projectIds: []});
      expect(initialRepos).toHaveLength(1);

      await client.pfs.deleteRepo({
        projectId: 'default',
        repo: {name: 'deleteRepo'},
      });

      const updatedRepos = await client.pfs.listRepo({projectIds: []});
      expect(updatedRepos).toHaveLength(0);
    });
  });
  describe('deleteRepos', () => {
    it('should delete all repos across projects', async () => {
      const client = await createSandbox('deleteRepos');
      await client.pfs.createRepo({
        projectId: 'default',
        repo: {name: 'a-repo'},
      });
      await client.pfs.createProject({name: 'project-2'});
      await client.pfs.createRepo({
        projectId: 'project-2',
        repo: {name: 'a-repo1'},
      });
      await client.pfs.createRepo({
        projectId: 'project-2',
        repo: {name: 'a-repo2'},
      });

      const initialRepos = await client.pfs.listRepo({
        projectIds: ['default', 'project-2'],
      });
      expect(initialRepos).toHaveLength(4);

      await client.pfs.deleteRepos({projectIds: ['default', 'project-2']});

      const updatedRepos = await client.pfs.listRepo({
        projectIds: ['default', 'project-2'],
      });
      expect(updatedRepos).toHaveLength(0);
    });
  });
  describe('diffFile', () => {
    it('should return a list of file diffs', async () => {
      const client = await createSandbox('diffFile');

      const commit1 = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'diffFile'}},
      });

      const fileClient1 = await client.pfs.modifyFile();
      await fileClient1
        .setCommit(commit1)
        .putFileFromFilepath(atatPngFilePath, 'at-at.png')
        .end();
      await client.pfs.finishCommit({projectId: 'default', commit: commit1});

      const commit2 = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'diffFile'}},
      });
      const fileClient2 = await client.pfs.modifyFile();
      await fileClient2
        .setCommit(commit2)
        .putFileFromFilepath(libertyPngFilePath, 'liberty.png')
        .end();
      await client.pfs.finishCommit({projectId: 'default', commit: commit2});

      const commitSet1 = await client.pfs.inspectCommitSet({
        commitSet: commit1,
      });

      expect(commitSet1[0].details?.sizeBytes).toBe(80588);

      const commitSet2 = await client.pfs.inspectCommitSet({
        commitSet: commit2,
      });

      expect(commitSet2[0].details?.sizeBytes).toBe(139232);

      const fileDiff1 = await client.pfs.diffFile({
        projectId: 'default',
        newFileObject: {
          commitId: commit2.id,
          path: '/',
          branch: {name: 'master', repo: {name: 'diffFile'}},
        },
      });
      expect(fileDiff1[1].newFile?.sizeBytes).toEqual(139232 - 80588);

      const fileDiff2 = await client.pfs.diffFile({
        projectId: 'default',
        newFileObject: {
          commitId: commit1.id,
          path: '/',
          branch: {name: 'master', repo: {name: 'diffFile'}},
        },
      });
      expect(fileDiff2[1].newFile?.sizeBytes).toBe(80588);
    });
  });

  describe('createProject', () => {
    it('should create a project', async () => {
      const client = await createSandbox('createProject');

      const initialProjects = await client.pfs.listProject();
      expect(initialProjects).toHaveLength(1);

      await client.pfs.createProject({name: 'name', description: 'desc'});

      const updatedProjects = await client.pfs.listProject();
      expect(updatedProjects).toHaveLength(2);
      expect(updatedProjects[1]).toEqual(
        expect.objectContaining({
          project: {
            name: 'name',
          },
          description: 'desc',
        }),
      );
    });

    it('should update a project', async () => {
      const client = await createSandbox('updateProject');

      const initialProjects = await client.pfs.listProject();
      expect(initialProjects).toHaveLength(1);
      expect(initialProjects[0]).toEqual(
        expect.objectContaining({
          project: {
            name: 'default',
          },
          description: '',
        }),
      );

      await client.pfs.createProject({
        name: 'default',
        description: 'desc',
        update: true,
      });

      const updatedProjects = await client.pfs.listProject();
      expect(updatedProjects[0]).toEqual(
        expect.objectContaining({
          project: {
            name: 'default',
          },
          description: 'desc',
        }),
      );
    });
  });
  describe('listProject', () => {
    it('should return a list of projects in the pachyderm cluster', async () => {
      const client = await createSandbox('listProject');
      const projects = await client.pfs.listProject();

      expect(projects).toEqual(
        expect.arrayContaining([
          {
            project: {name: 'default'},
            description: '',
            authInfo: undefined,
            // Example: createdAt: { seconds: 1691596017, nanos: 248610000 }
            createdAt: {seconds: expect.any(Number), nanos: expect.any(Number)},
          },
        ]),
      );
    });
  });
  describe('inspectProject', () => {
    it('should return information about the specified', async () => {
      const client = await createSandbox('inspectProject');
      const projectInfo = await client.pfs.inspectProject('default');

      expect(projectInfo).toEqual(
        expect.objectContaining({
          project: {name: 'default'},
          description: '',
          authInfo: undefined,
          // Example: createdAt: { seconds: 1691596017, nanos: 248610000 }
          createdAt: {seconds: expect.any(Number), nanos: expect.any(Number)},
        }),
      );
    });
  });
  describe('deleteProject', () => {
    it('should delete a project', async () => {
      const client = await createSandbox('deleteProject');
      await client.pfs.deleteAll();
      expect(await client.pfs.listProject()).toHaveLength(1);

      await client.pfs.createProject({name: 'test-delete-project'});
      expect(await client.pfs.listProject()).toHaveLength(2);

      await client.pfs.deleteProject({projectId: 'test-delete-project'});
      expect(await client.pfs.listProject()).toHaveLength(1);
      expect(await client.pfs.inspectProject('default')).toBeTruthy();
    });
  });
});
