import path from 'path';

import apiClientRequestWrapper from '../../../../client';

const libertyPngFilePath = path.resolve(
  __dirname,
  '../../../../../../../etc/testing/files/liberty.png',
);
const atatPngFilePath = path.resolve(
  __dirname,
  '../../../../../../../etc/testing/files/AT-AT.png',
);

jest.setTimeout(30_000);

async function retry<T>(maxRetries: number, fn: () => Promise<T>): Promise<T> {
  try {
    return await fn();
  } catch (err) {
    if (maxRetries <= 0) {
      throw err;
    }
    return retry(maxRetries - 1, fn);
  }
}

describe('ModifyFile', () => {
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

  describe('putFileFromURL', () => {
    it('should add a file from a URL to a repo', async () => {
      const client = await createSandbox('putFileFromURL');
      const commit = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'putFileFromURL'}},
      });

      const initialFiles = await client.pfs.listFile({
        projectId: 'default',
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'putFileFromURL'}},
      });
      expect(initialFiles).toHaveLength(0);

      const fileClient = await client.pfs.modifyFile();
      await fileClient
        .setCommit(commit)
        .putFileFromFilepath(atatPngFilePath, 'at-at.png')
        .putFileFromFilepath(libertyPngFilePath, 'liberty.png')
        .end();

      await client.pfs.finishCommit({projectId: 'default', commit});

      const files = await client.pfs.listFile({
        projectId: 'default',
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'putFileFromURL'}},
      });

      expect(files).toHaveLength(2);
    });
  });

  describe('putFileFromBytes', () => {
    it('should add a file in byte format to a repo', async () => {
      const client = await createSandbox('putFileFromBytes');
      const commit = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'putFileFromBytes'}},
      });

      const initialFiles = await client.pfs.listFile({
        projectId: 'default',
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'putFileFromBytes'}},
      });
      expect(initialFiles).toHaveLength(0);

      const fileClient = await client.pfs.modifyFile();
      await fileClient
        .setCommit(commit)
        .putFileFromBytes('test.dat', Buffer.from('data'))
        .putFileFromBytes('test2.dat', Buffer.from('data'))
        .end();
      await client.pfs.finishCommit({projectId: 'default', commit});

      const files = await client.pfs.listFile({
        projectId: 'default',
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'putFileFromBytes'}},
      });
      expect(files).toHaveLength(2);
    }, 60_000);
  });

  describe('putFileFromFilePath', () => {
    it('should add a file from a local path to the repo', async () => {
      const client = await createSandbox('putFileFromFilePath');

      const commit = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'putFileFromFilePath'}},
      });

      const initialFiles = await client.pfs.listFile({
        projectId: 'default',
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'putFileFromFilePath'}},
      });
      expect(initialFiles).toHaveLength(0);

      const fileClient = await client.pfs.modifyFile();

      await fileClient
        .setCommit(commit)
        .putFileFromFilepath(
          path.join(
            __dirname,
            '../../../../../../../etc/testing/files/8MN9Kg0.jpg',
          ),
          '/8MN9Kg0.jpg',
        )
        .putFileFromFilepath(
          path.join(
            __dirname,
            '../../../../../../../etc/testing/files/46Q8nDz.jpg',
          ),
          '/46Q8nDz.jpg',
        )
        .end();

      await client.pfs.finishCommit({projectId: 'default', commit});

      await retry(3, async () => {
        const files = await client.pfs.listFile({
          projectId: 'default',
          commitId: commit.id,
          branch: {name: 'master', repo: {name: 'putFileFromFilePath'}},
        });
        expect(files).toHaveLength(2);
      });
    }, 65_000);
  });

  describe('deleteFile', () => {
    it('should delete a file by path', async () => {
      const client = await createSandbox('deleteFile');
      const commit = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'deleteFile'}},
      });

      const fileClient = await client.pfs.modifyFile();

      await fileClient
        .setCommit(commit)
        .putFileFromBytes('test.dat', Buffer.from('data'))
        .putFileFromBytes('test2.dat', Buffer.from('data'))
        .end();

      await client.pfs.finishCommit({projectId: 'default', commit});

      await retry(3, async () => {
        const files = await client.pfs.listFile({
          projectId: 'default',
          commitId: commit.id,
          branch: {name: 'master', repo: {name: 'deleteFile'}},
        });
        expect(files).toHaveLength(2);
      });

      const deleteCommit = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'deleteFile'}},
      });

      const fileClient2 = await client.pfs.modifyFile();

      await fileClient2
        .setCommit(deleteCommit)
        .deleteFile('test.dat')
        .deleteFile('test2.dat')
        .end();

      await client.pfs.finishCommit({
        projectId: 'default',
        commit: deleteCommit,
      });

      const postDeleteFiles = await client.pfs.listFile({
        projectId: 'default',
        commitId: deleteCommit.id,
        branch: {name: 'master', repo: {name: 'deleteFile'}},
      });
      expect(postDeleteFiles).toHaveLength(0);
    }, 65_000);
  });

  describe('deleteFiles', () => {
    it('should delete files by path', async () => {
      const client = await createSandbox('deleteFiles');
      const commit = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'deleteFiles'}},
      });

      const fileClient = await client.pfs.modifyFile();

      await fileClient
        .setCommit(commit)
        .putFileFromBytes('test.dat', Buffer.from('data'))
        .putFileFromBytes('test2.dat', Buffer.from('data'))
        .end();

      await client.pfs.finishCommit({projectId: 'default', commit});

      const files = await client.pfs.listFile({
        projectId: 'default',
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'deleteFiles'}},
      });
      expect(files).toHaveLength(2);

      const deleteCommit = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'deleteFiles'}},
      });

      const fileClient2 = await client.pfs.modifyFile();

      await fileClient2
        .setCommit(deleteCommit)
        .deleteFiles(['test.dat', 'test2.dat'])
        .end();

      await client.pfs.finishCommit({
        projectId: 'default',
        commit: deleteCommit,
      });

      const postDeleteFiles = await client.pfs.listFile({
        projectId: 'default',
        commitId: deleteCommit.id,
        branch: {name: 'master', repo: {name: 'deleteFiles'}},
      });
      expect(postDeleteFiles).toHaveLength(0);
    }, 60_000);
  });

  describe('autoCommit', () => {
    it('should be able to use auto commits', async () => {
      const client = await createSandbox('putFileFromFilePath');

      const commits = await client.pfs.listCommit({
        projectId: 'default',
        repo: {name: 'putFileFromFilePath'},
      });
      expect(commits).toHaveLength(0);

      const fileClient = await client.pfs.modifyFile();

      await fileClient
        .autoCommit({name: 'master', repo: {name: 'putFileFromFilePath'}})
        .putFileFromFilepath(
          path.join(
            __dirname,
            '../../../../../../../etc/testing/files/8MN9Kg0.jpg',
          ),
          '/8MN9Kg0.jpg',
        )
        .putFileFromFilepath(
          path.join(
            __dirname,
            '../../../../../../../etc/testing/files/46Q8nDz.jpg',
          ),
          '/46Q8nDz.jpg',
        )
        .end();

      const newCommits = await client.pfs.listCommit({
        projectId: 'default',
        repo: {name: 'putFileFromFilePath'},
      });
      expect(newCommits).toHaveLength(1);
      const files = await client.pfs.listFile({
        projectId: 'default',
        commitId: '',
        branch: {name: 'master', repo: {name: 'putFileFromFilePath'}},
      });
      expect(files).toHaveLength(2);
    });
  });
});
