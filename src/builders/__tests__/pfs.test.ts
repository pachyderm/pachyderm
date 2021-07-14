import {
  commitFromObject,
  createRepoRequestFromObject,
  deleteRepoRequestFromObject,
  fileFromObject,
  fileInfoFromObject,
  inspectFileRequestFromObject,
  repoFromObject,
  triggerFromObject,
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

  it('should inspect a file from an object with defaults', () => {
    const file = fileFromObject({
      branch: {name: 'master', repo: {name: 'neato'}},
    });
    const inspectedFile = inspectFileRequestFromObject({file})

    expect(inspectedFile.getFile()?.getCommit()?.getId()).toBe('master');
    expect(inspectedFile.getFile()?.getCommit()?.getBranch()?.getRepo()?.getName()).toBe('neato');
    expect(inspectedFile.getFile()?.getPath()).toBe('/');
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

  it('should create createRepoRequest from an object with defaults', () => {
    const createRepoRequest = createRepoRequestFromObject({
      repo: {name: 'test'},
    });

    expect(createRepoRequest.getRepo()?.getName()).toBe('test');
    expect(createRepoRequest.getDescription()).toBe('');
    expect(createRepoRequest.getUpdate()).toBe(false);
  });

  it('should create createRepoRequest from an object', () => {
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
      force: true
    });

    expect(deleteRepoRequest.getRepo()?.getName()).toBe('test');
    expect(deleteRepoRequest.getForce()).toBe(true);
  });
});
