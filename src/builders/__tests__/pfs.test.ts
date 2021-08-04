import {CommitState, OriginKind} from '@pachyderm/proto/pb/pfs/pfs_pb';

import {
  commitFromObject,
  createRepoRequestFromObject,
  deleteRepoRequestFromObject,
  startCommitRequestFromObject,
  finishCommitRequestFromObject,
  inspectCommitRequestFromObject,
  listCommitRequestFromObject,
  subscribeCommitRequestFromObject,
  commitSetFromObject,
  fileFromObject,
  fileInfoFromObject,
  inspectCommitSetRequestFromObject,
  createBranchRequestFromObject,
  listBranchRequestFromObject,
  deleteBranchRequestFromObject,
  repoFromObject,
  triggerFromObject,
  commitInfoFromObject,
} from '../pfs';

describe('grpc/builders/pfs', () => {
  it('should create File from an object', () => {
    const file = fileFromObject({
      commitId: '1234567890',
      path: '/assets',
      branch: {name: 'master', repo: {name: 'neato'}},
    });

    expect(file.getCommit()?.getId()).toBe('1234567890');
    expect(file.getCommit()?.getBranch()?.getRepo()?.getName()).toBe('neato');
    expect(file.getPath()).toBe('/assets');
  });

  it('should create File from an object with defaults', () => {
    const file = fileFromObject({
      branch: {name: 'master', repo: {name: 'neato'}},
    });

    expect(file.getCommit()?.getId()).toBe('master');
    expect(file.getCommit()?.getBranch()?.getRepo()?.getName()).toBe('neato');
    expect(file.getPath()).toBe('/');
  });

  it('should create FileInfo from an object', () => {
    const fileInfo = fileInfoFromObject({
      committed: {
        seconds: 1615922718,
        nanos: 449796812,
      },
      file: {
        commitId: '1234567890',
        path: '/assets',
        branch: {name: 'master', repo: {name: 'neato'}},
      },
      fileType: 2,
      hash: 'abcde12345',
      sizeBytes: 123,
    });

    expect(fileInfo.getFile()?.getCommit()?.getId()).toBe('1234567890');
    expect(
      fileInfo.getFile()?.getCommit()?.getBranch()?.getRepo()?.getName(),
    ).toBe('neato');
    expect(fileInfo.getFile()?.getPath()).toBe('/assets');
    expect(fileInfo.getFileType()).toBe(2);
    expect(fileInfo.getHash()).toBe('abcde12345');
    expect(fileInfo.getSizeBytes()).toBe(123);
    expect(fileInfo.getCommitted()?.getSeconds()).toBe(1615922718);
    expect(fileInfo.getCommitted()?.getNanos()).toBe(449796812);
  });

  it('should create FileInfo from an object with missing timestamp', () => {
    const fileInfo = fileInfoFromObject({
      committed: undefined,
      file: {
        commitId: '1234567890',
        path: '/assets',
        branch: {name: 'master', repo: {name: 'neato'}},
      },
      fileType: 2,
      hash: 'abcde12345',
      sizeBytes: 123,
    });

    expect(fileInfo.getFile()?.getCommit()?.getId()).toBe('1234567890');
    expect(
      fileInfo.getFile()?.getCommit()?.getBranch()?.getRepo()?.getName(),
    ).toBe('neato');
    expect(fileInfo.getFile()?.getPath()).toBe('/assets');
    expect(fileInfo.getFileType()).toBe(2);
    expect(fileInfo.getHash()).toBe('abcde12345');
    expect(fileInfo.getSizeBytes()).toBe(123);
    expect(fileInfo.getCommitted()?.getSeconds()).toBe(undefined);
    expect(fileInfo.getCommitted()?.getNanos()).toBe(undefined);
  });

  it('should create Trigger from an object', () => {
    const trigger = triggerFromObject({
      branch: 'master',
      all: true,
      cronSpec: '@every 10s',
      size: 'big',
      commits: 12,
    });

    expect(trigger.getBranch()).toBe('master');
    expect(trigger.getAll()).toBe(true);
    expect(trigger.getCronSpec()).toBe('@every 10s');
    expect(trigger.getSize()).toBe('big');
    expect(trigger.getCommits()).toBe(12);
  });

  it('should create Repo from an object', () => {
    const repo = repoFromObject({
      name: '__spec__',
    });

    expect(repo.getName()).toBe('__spec__');
  });

  it('should create Commit from an object', () => {
    const commit = commitFromObject({
      branch: {name: 'master', repo: {name: '__spec__'}},
      id: '4af40d34a0384f23a5b98d3bd7eaece1',
    });

    expect(commit.getBranch()?.getRepo()?.getName()).toBe('__spec__');
    expect(commit.getId()).toBe('4af40d34a0384f23a5b98d3bd7eaece1');
  });

  it('should create commitInfo from an object with defaults', () => {
    const commitInfo = commitInfoFromObject({
      commit: {
        branch: {name: 'master', repo: {name: '__spec__'}},
        id: '4af40d34a0384f23a5b98d3bd7eaece1',
      },
    });

    expect(commitInfo.getCommit()?.getBranch()?.getName()).toBe('master');
    expect(commitInfo.getCommit()?.getBranch()?.getRepo()?.getName()).toBe(
      '__spec__',
    );
    expect(commitInfo.getCommit()?.getId()).toBe(
      '4af40d34a0384f23a5b98d3bd7eaece1',
    );

    expect(commitInfo.getDescription()).toBe('');
    expect(commitInfo.getDetails()?.getSizeBytes()).toBe(0);
    expect(commitInfo.getStarted()?.getSeconds()).toBe(undefined);
    expect(commitInfo.getStarted()?.getNanos()).toBe(undefined);
    expect(commitInfo.getFinishing()?.getSeconds()).toBe(undefined);
    expect(commitInfo.getFinishing()?.getNanos()).toBe(undefined);
    expect(commitInfo.getFinished()?.getSeconds()).toBe(undefined);
    expect(commitInfo.getFinished()?.getNanos()).toBe(undefined);
    expect(commitInfo.getSizeBytesUpperBound()).toBe(0);
  });

  it('should create commitInfo from an object', () => {
    const commitInfo = commitInfoFromObject({
      commit: {
        branch: {name: 'master', repo: {name: '__spec__'}},
        id: '4af40d34a0384f23a5b98d3bd7eaece1',
      },
      description: 'this is a description',
      sizeBytes: 342,
      started: {
        seconds: 1615922000,
        nanos: 449796000,
      },
      finishing: {
        seconds: 1615922010,
        nanos: 449796010,
      },
      finished: {
        seconds: 1615922001,
        nanos: 449796001,
      },
      sizeBytesUpperBound: 200,
    });

    expect(commitInfo.getCommit()?.getBranch()?.getName()).toBe('master');
    expect(commitInfo.getCommit()?.getBranch()?.getRepo()?.getName()).toBe(
      '__spec__',
    );
    expect(commitInfo.getCommit()?.getId()).toBe(
      '4af40d34a0384f23a5b98d3bd7eaece1',
    );

    expect(commitInfo.getDescription()).toBe('this is a description');
    expect(commitInfo.getDetails()?.getSizeBytes()).toBe(342);
    expect(commitInfo.getStarted()?.getSeconds()).toBe(1615922000);
    expect(commitInfo.getStarted()?.getNanos()).toBe(449796000);
    expect(commitInfo.getFinishing()?.getSeconds()).toBe(1615922010);
    expect(commitInfo.getFinishing()?.getNanos()).toBe(449796010);
    expect(commitInfo.getFinished()?.getSeconds()).toBe(1615922001);
    expect(commitInfo.getFinished()?.getNanos()).toBe(449796001);
    expect(commitInfo.getSizeBytesUpperBound()).toBe(200);
  });

  it('should create CommitSet from an object', () => {
    const commitSet = commitSetFromObject({
      id: '4af40d34a0384f23a5b98d3bd7eaece1',
    });

    expect(commitSet.getId()).toBe('4af40d34a0384f23a5b98d3bd7eaece1');
  });

  it('should create a startCommitRequest from an object with defaults', () => {
    const startCommitRequest = startCommitRequestFromObject({
      branch: {name: 'master', repo: {name: '__spec__'}},
    });

    expect(startCommitRequest.getBranch()?.getName()).toBe('master');
    expect(startCommitRequest.getBranch()?.getRepo()?.getName()).toBe(
      '__spec__',
    );
    expect(startCommitRequest.getDescription()).toBe('');
  });

  it('should create a startCommitRequest from an object with description', () => {
    const startCommitRequest = startCommitRequestFromObject({
      branch: {name: 'master', repo: {name: '__spec__'}},
      description: 'neato',
    });

    expect(startCommitRequest.getBranch()?.getName()).toBe('master');
    expect(startCommitRequest.getBranch()?.getRepo()?.getName()).toBe(
      '__spec__',
    );
    expect(startCommitRequest.getDescription()).toBe('neato');
  });

  it('should create a startCommitRequest from an object with parent commit', () => {
    const startCommitRequest = startCommitRequestFromObject({
      branch: {name: 'master', repo: {name: '__spec__'}},
      description: 'neato',
      parent: {
        branch: {name: 'master', repo: {name: '__spec__'}},
        id: '4af40d34a0384f23a5b98d3bd7eaece1',
      },
    });

    expect(startCommitRequest.getBranch()?.getName()).toBe('master');
    expect(startCommitRequest.getBranch()?.getRepo()?.getName()).toBe(
      '__spec__',
    );
    expect(startCommitRequest.getDescription()).toBe('neato');
    expect(startCommitRequest.getParent()?.getBranch()?.getName()).toBe(
      'master',
    );
    expect(
      startCommitRequest.getParent()?.getBranch()?.getRepo()?.getName(),
    ).toBe('__spec__');
    expect(startCommitRequest.getParent()?.getId()).toBe(
      '4af40d34a0384f23a5b98d3bd7eaece1',
    );
  });

  it('should create a finishCommitRequest from an object with defaults', () => {
    const finishCommitRequest = finishCommitRequestFromObject({
      commit: {
        branch: {name: 'master', repo: {name: '__spec__'}},
        id: '4af40d34a0384f23a5b98d3bd7eaece1',
      },
    });

    expect(finishCommitRequest.getCommit()?.getBranch()?.getName()).toBe(
      'master',
    );
    expect(
      finishCommitRequest.getCommit()?.getBranch()?.getRepo()?.getName(),
    ).toBe('__spec__');
    expect(finishCommitRequest.getCommit()?.getId()).toBe(
      '4af40d34a0384f23a5b98d3bd7eaece1',
    );
  });

  it('should create a finishCommitRequest from an object with description', () => {
    const finishCommitRequest = finishCommitRequestFromObject({
      commit: {
        branch: {name: 'master', repo: {name: '__spec__'}},
        id: '4af40d34a0384f23a5b98d3bd7eaece1',
      },
      description: 'neato',
    });

    expect(finishCommitRequest.getCommit()?.getBranch()?.getName()).toBe(
      'master',
    );
    expect(
      finishCommitRequest.getCommit()?.getBranch()?.getRepo()?.getName(),
    ).toBe('__spec__');
    expect(finishCommitRequest.getCommit()?.getId()).toBe(
      '4af40d34a0384f23a5b98d3bd7eaece1',
    );
    expect(finishCommitRequest.getDescription()).toBe('neato');
  });

  it('should create a finishCommitRequest from an object with force enabled', () => {
    const finishCommitRequest = finishCommitRequestFromObject({
      commit: {
        branch: {name: 'master', repo: {name: '__spec__'}},
        id: '4af40d34a0384f23a5b98d3bd7eaece1',
      },
      description: 'neato',
      force: true,
    });

    expect(finishCommitRequest.getCommit()?.getBranch()?.getName()).toBe(
      'master',
    );
    expect(
      finishCommitRequest.getCommit()?.getBranch()?.getRepo()?.getName(),
    ).toBe('__spec__');
    expect(finishCommitRequest.getCommit()?.getId()).toBe(
      '4af40d34a0384f23a5b98d3bd7eaece1',
    );
    expect(finishCommitRequest.getDescription()).toBe('neato');
    expect(finishCommitRequest.getForce()).toBe(true);
  });

  it('should create a finishCommitRequest from an object with an error', () => {
    const finishCommitRequest = finishCommitRequestFromObject({
      commit: {
        branch: {name: 'master', repo: {name: '__spec__'}},
        id: '4af40d34a0384f23a5b98d3bd7eaece1',
      },
      description: 'neato',
      force: true,
      error: 'error',
    });

    expect(finishCommitRequest.getCommit()?.getBranch()?.getName()).toBe(
      'master',
    );
    expect(
      finishCommitRequest.getCommit()?.getBranch()?.getRepo()?.getName(),
    ).toBe('__spec__');
    expect(finishCommitRequest.getCommit()?.getId()).toBe(
      '4af40d34a0384f23a5b98d3bd7eaece1',
    );
    expect(finishCommitRequest.getDescription()).toBe('neato');
    expect(finishCommitRequest.getForce()).toBe(true);
    expect(finishCommitRequest.getError()).toBe('error');
  });

  it('should create a inspectCommitRequest from an object with wait and commit', () => {
    const inspectCommitRequest = inspectCommitRequestFromObject({
      wait: CommitState.READY,
      commit: {
        branch: {name: 'master', repo: {name: '__spec__'}},
        id: '4af40d34a0384f23a5b98d3bd7eaece1',
      },
    });

    expect(inspectCommitRequest.getWait()).toBe(2); // <- ready state's enum index is 2
    expect(inspectCommitRequest.getCommit()?.getBranch()?.getName()).toBe(
      'master',
    );
    expect(
      inspectCommitRequest.getCommit()?.getBranch()?.getRepo()?.getName(),
    ).toBe('__spec__');
    expect(inspectCommitRequest.getCommit()?.getId()).toBe(
      '4af40d34a0384f23a5b98d3bd7eaece1',
    );
  });

  it('should create a listCommitRequest from an object with defaults', () => {
    const listCommitRequest = listCommitRequestFromObject({
      repo: {
        name: '__spec__',
      },
    });

    expect(listCommitRequest.getRepo()?.getName()).toBe('__spec__');
    expect(listCommitRequest.getRepo()?.getType()).toBe('user');
    expect(listCommitRequest.getAll()).toBe(true);
    expect(listCommitRequest.getReverse()).toBe(false);
  });

  it('should create a listCommitRequest from an object with from', () => {
    const listCommitRequest = listCommitRequestFromObject({
      repo: {
        name: '__spec__',
      },
      from: {
        branch: {name: 'master', repo: {name: '__spec__'}},
        id: '4af40d34a0384f23a5b98d3bd7eaece1',
      },
    });

    expect(listCommitRequest.getRepo()?.getName()).toBe('__spec__');
    expect(listCommitRequest.getRepo()?.getType()).toBe('user');
    expect(listCommitRequest.getAll()).toBe(true);
    expect(listCommitRequest.getReverse()).toBe(false);
    expect(listCommitRequest.getFrom()?.getBranch()?.getName()).toBe('master');
    expect(listCommitRequest.getFrom()?.getBranch()?.getRepo()?.getName()).toBe(
      '__spec__',
    );
    expect(listCommitRequest.getFrom()?.getId()).toBe(
      '4af40d34a0384f23a5b98d3bd7eaece1',
    );
  });

  it('should create a listCommitRequest from an object with to', () => {
    const listCommitRequest = listCommitRequestFromObject({
      repo: {
        name: '__spec__',
      },
      to: {
        branch: {name: 'master', repo: {name: '__spec__'}},
        id: '4af40d34a0384f23a5b98d3bd7eaece1',
      },
    });

    expect(listCommitRequest.getRepo()?.getName()).toBe('__spec__');
    expect(listCommitRequest.getRepo()?.getType()).toBe('user');
    expect(listCommitRequest.getAll()).toBe(true);
    expect(listCommitRequest.getReverse()).toBe(false);
    expect(listCommitRequest.getTo()?.getBranch()?.getName()).toBe('master');
    expect(listCommitRequest.getTo()?.getBranch()?.getRepo()?.getName()).toBe(
      '__spec__',
    );
    expect(listCommitRequest.getTo()?.getId()).toBe(
      '4af40d34a0384f23a5b98d3bd7eaece1',
    );
  });

  it('should create a listCommitRequest from an object with number', () => {
    const listCommitRequest = listCommitRequestFromObject({
      repo: {
        name: '__spec__',
      },
      number: 1,
    });

    expect(listCommitRequest.getRepo()?.getName()).toBe('__spec__');
    expect(listCommitRequest.getRepo()?.getType()).toBe('user');
    expect(listCommitRequest.getAll()).toBe(true);
    expect(listCommitRequest.getReverse()).toBe(false);
    expect(listCommitRequest.getNumber()).toBe(1);
  });

  it('should create a listCommitRequest from an object with originKind', () => {
    const listCommitRequest = listCommitRequestFromObject({
      repo: {
        name: '__spec__',
      },
      originKind: OriginKind.USER,
    });

    expect(listCommitRequest.getRepo()?.getName()).toBe('__spec__');
    expect(listCommitRequest.getRepo()?.getType()).toBe('user');
    expect(listCommitRequest.getAll()).toBe(true);
    expect(listCommitRequest.getReverse()).toBe(false);
    expect(listCommitRequest.getOriginKind()).toBe(1);
  });

  it('should create a listCommitRequest from an object with setAll to false', () => {
    const listCommitRequest = listCommitRequestFromObject({
      repo: {
        name: '__spec__',
      },
      all: false,
    });

    expect(listCommitRequest.getRepo()?.getName()).toBe('__spec__');
    expect(listCommitRequest.getRepo()?.getType()).toBe('user');
    expect(listCommitRequest.getAll()).toBe(false);
    expect(listCommitRequest.getReverse()).toBe(false);
  });

  it('should create a listCommitRequest from an object with reverse to true', () => {
    const listCommitRequest = listCommitRequestFromObject({
      repo: {
        name: '__spec__',
      },
      reverse: true,
    });

    expect(listCommitRequest.getRepo()?.getName()).toBe('__spec__');
    expect(listCommitRequest.getRepo()?.getType()).toBe('user');
    expect(listCommitRequest.getAll()).toBe(true);
    expect(listCommitRequest.getReverse()).toBe(true);
  });

  it('should create a subscribeCommitRequest from an object with defaults', () => {
    const subscribeCommitRequest = subscribeCommitRequestFromObject({
      repo: {
        name: '__spec__',
      },
    });

    expect(subscribeCommitRequest.getRepo()?.getName()).toBe('__spec__');
    expect(subscribeCommitRequest.getRepo()?.getType()).toBe('user');
    expect(subscribeCommitRequest.getAll()).toBe(true);
  });

  it('should create a subscribeCommitRequest from an object with from commit set', () => {
    const subscribeCommitRequest = subscribeCommitRequestFromObject({
      repo: {
        name: '__spec__',
      },
      from: {
        branch: {name: 'master', repo: {name: '__spec__'}},
        id: '4af40d34a0384f23a5b98d3bd7eaece1',
      },
    });

    expect(subscribeCommitRequest.getRepo()?.getName()).toBe('__spec__');
    expect(subscribeCommitRequest.getRepo()?.getType()).toBe('user');
    expect(subscribeCommitRequest.getAll()).toBe(true);
    expect(subscribeCommitRequest.getFrom()?.getBranch()?.getName()).toBe(
      'master',
    );
    expect(
      subscribeCommitRequest.getFrom()?.getBranch()?.getRepo()?.getName(),
    ).toBe('__spec__');
    expect(
      subscribeCommitRequest.getFrom()?.getBranch()?.getRepo()?.getType(),
    ).toBe('user');
    expect(subscribeCommitRequest.getFrom()?.getId()).toBe(
      '4af40d34a0384f23a5b98d3bd7eaece1',
    );
  });

  it('should create a subscribeCommitRequest from an object with branch set', () => {
    const subscribeCommitRequest = subscribeCommitRequestFromObject({
      repo: {
        name: '__spec__',
      },
      branch: 'master',
    });

    expect(subscribeCommitRequest.getRepo()?.getName()).toBe('__spec__');
    expect(subscribeCommitRequest.getRepo()?.getType()).toBe('user');
    expect(subscribeCommitRequest.getAll()).toBe(true);
    expect(subscribeCommitRequest.getBranch()).toBe('master');
  });

  it('should create a subscribeCommitRequest from an object with state set', () => {
    const subscribeCommitRequest = subscribeCommitRequestFromObject({
      repo: {
        name: '__spec__',
      },
      state: CommitState.FINISHED,
    });

    expect(subscribeCommitRequest.getRepo()?.getName()).toBe('__spec__');
    expect(subscribeCommitRequest.getRepo()?.getType()).toBe('user');
    expect(subscribeCommitRequest.getAll()).toBe(true);
    expect(subscribeCommitRequest.getState()).toBe(4);
  });

  it('should create a subscribeCommitRequest from an object with originKind set', () => {
    const subscribeCommitRequest = subscribeCommitRequestFromObject({
      repo: {
        name: '__spec__',
      },
      originKind: OriginKind.AUTO,
    });

    expect(subscribeCommitRequest.getRepo()?.getName()).toBe('__spec__');
    expect(subscribeCommitRequest.getRepo()?.getType()).toBe('user');
    expect(subscribeCommitRequest.getAll()).toBe(true);
    expect(subscribeCommitRequest.getOriginKind()).toBe(2);
  });

  it('should create a subscribeCommitRequest from an object with all set to false', () => {
    const subscribeCommitRequest = subscribeCommitRequestFromObject({
      repo: {
        name: '__spec__',
      },
      all: false,
    });

    expect(subscribeCommitRequest.getRepo()?.getName()).toBe('__spec__');
    expect(subscribeCommitRequest.getRepo()?.getType()).toBe('user');
    expect(subscribeCommitRequest.getAll()).toBe(false);
  });

  it('should create inspectCommitSetRequest from an object with defaults to wait for commits to finish', () => {
    const commitSet = inspectCommitSetRequestFromObject({
      commitSet: {id: '4af40d34a0384f23a5b98d3bd7eaece1'},
    });

    expect(commitSet.getCommitSet()?.getId()).toBe(
      '4af40d34a0384f23a5b98d3bd7eaece1',
    );
    expect(commitSet.getWait()).toBe(true);
  });

  it('should create inspectCommitSetRequest from an object wait set to false', () => {
    const commitSet = inspectCommitSetRequestFromObject({
      commitSet: {id: '4af40d34a0384f23a5b98d3bd7eaece1'},
      wait: false,
    });

    expect(commitSet.getCommitSet()?.getId()).toBe(
      '4af40d34a0384f23a5b98d3bd7eaece1',
    );
    expect(commitSet.getWait()).toBe(false);
  });

  it('should create createBranchRequest from an object with defaults', () => {
    const createBranchRequest = createBranchRequestFromObject({
      head: {
        branch: {name: 'master', repo: {name: '__spec__'}},
        id: '4af40d34a0384f23a5b98d3bd7eaece1',
      },
      branch: {
        name: 'staging',
        repo: {name: '__spec__'},
      },
      trigger: {
        branch: 'master',
        all: true,
        cronSpec: '@every 10s',
        size: '1MB',
        commits: 12,
      },
      provenance: [],
      newCommitSet: false,
    });
    expect(createBranchRequest.getHead()?.getBranch()?.getName()).toBe(
      'master',
    );
    expect(
      createBranchRequest.getHead()?.getBranch()?.getRepo()?.getName(),
    ).toBe('__spec__');
    expect(createBranchRequest.getHead()?.getId()).toBe(
      '4af40d34a0384f23a5b98d3bd7eaece1',
    );
    expect(createBranchRequest.getBranch()?.getName()).toBe('staging');
    expect(createBranchRequest.getBranch()?.getRepo()?.getName()).toBe(
      '__spec__',
    );
    expect(createBranchRequest.getProvenanceList()).toStrictEqual([]);
    expect(createBranchRequest.getTrigger()?.getBranch()).toBe('master');
    expect(createBranchRequest.getTrigger()?.getAll()).toBe(true);
    expect(createBranchRequest.getTrigger()?.getCronSpec()).toBe('@every 10s');
    expect(createBranchRequest.getTrigger()?.getSize()).toBe('1MB');
    expect(createBranchRequest.getTrigger()?.getCommits()).toBe(12);
    expect(createBranchRequest.getNewCommitSet()).toBe(false);
  });

  it('should create listBranchRequest from an object with defaults reverse to false', () => {
    const listBranchRequest = listBranchRequestFromObject({
      repo: {name: 'test'},
    });
    expect(listBranchRequest.getRepo()?.getName()).toBe('test');
    expect(listBranchRequest.getReverse()).toBe(false);
  });

  it('should create listBranchRequest from an object with reverse set to true', () => {
    const listBranchRequest = listBranchRequestFromObject({
      repo: {name: 'test'},
      reverse: true,
    });
    expect(listBranchRequest.getRepo()?.getName()).toBe('test');
    expect(listBranchRequest.getReverse()).toBe(true);
  });

  it('should create deleteBranchRequest from an object without force by default', () => {
    const deleteBranchRequest = deleteBranchRequestFromObject({
      branch: {name: 'master', repo: {name: 'test'}},
    });
    expect(deleteBranchRequest.getBranch()?.getName()).toBe('master');
    expect(deleteBranchRequest.getBranch()?.getRepo()?.getName()).toBe('test');
    expect(deleteBranchRequest.getForce()).toBe(false);
  });

  it('should create deleteBranchRequest from an object without force by default', () => {
    const deleteBranchRequest = deleteBranchRequestFromObject({
      branch: {name: 'master', repo: {name: 'test'}},
      force: true,
    });
    expect(deleteBranchRequest.getBranch()?.getName()).toBe('master');
    expect(deleteBranchRequest.getBranch()?.getRepo()?.getName()).toBe('test');
    expect(deleteBranchRequest.getForce()).toBe(true);
  });

  it('should create CreateRepoRequest from an object with defaults', () => {
    const createRepoRequest = createRepoRequestFromObject({
      repo: {name: 'test'},
    });

    expect(createRepoRequest.getRepo()?.getName()).toBe('test');
    expect(createRepoRequest.getDescription()).toBe('');
    expect(createRepoRequest.getUpdate()).toBe(false);
  });

  it('should create CreateRepoRequest from an object', () => {
    const createRepoRequest = createRepoRequestFromObject({
      repo: {name: 'test'},
      description: 'this is a discription.',
      update: true,
    });

    expect(createRepoRequest.getRepo()?.getName()).toBe('test');
    expect(createRepoRequest.getDescription()).toBe('this is a discription.');
    expect(createRepoRequest.getUpdate()).toBe(true);
  });

  it('should create deleteRepoRequest from an object without force by default', () => {
    const deleteRepoRequest = deleteRepoRequestFromObject({
      repo: {name: 'test'},
    });

    expect(deleteRepoRequest.getRepo()?.getName()).toBe('test');
    expect(deleteRepoRequest.getForce()).toBe(false);
  });

  it('should create deleteRepoRequest from an object with force', () => {
    const deleteRepoRequest = deleteRepoRequestFromObject({
      repo: {name: 'test'},
      force: true,
    });

    expect(deleteRepoRequest.getRepo()?.getName()).toBe('test');
    expect(deleteRepoRequest.getForce()).toBe(true);
  });
});
