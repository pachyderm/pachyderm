import path from 'path';

import client from '../../client';

describe('ModifyFile', () => {
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

  describe('putFileFromURL', () => {
    it('should add a file from a URL to a repo', async () => {
      const client = await createSandbox('putFileFromURL');
      const commit = await client.pfs().startCommit({
        branch: {name: 'master', repo: {name: 'putFileFromURL'}},
      });

      const initialFiles = await client.pfs().listFile({
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'putFileFromURL'}},
      });
      expect(initialFiles).toHaveLength(0);

      await client
        .modifyFile()
        .setCommit(commit)
        .putFileFromURL('at-at.png', 'http://imgur.com/8MN9Kg0.png')
        .end();

      await client.pfs().finishCommit({commit});

      const files = await client.pfs().listFile({
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'putFileFromURL'}},
      });

      expect(files).toHaveLength(1);
    });
  });

  describe('putFileFromBytes', () => {
    it('should add a file in byte format to a repo', async () => {
      const client = await createSandbox('putFileFromBytes');
      const commit = await client.pfs().startCommit({
        branch: {name: 'master', repo: {name: 'putFileFromBytes'}},
      });

      const initialFiles = await client.pfs().listFile({
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'putFileFromBytes'}},
      });
      expect(initialFiles).toHaveLength(0);

      await client
        .modifyFile()
        .setCommit(commit)
        .putFileFromBytes('test.dat', Buffer.from('data'))
        .end();
      await client.pfs().finishCommit({commit});

      const files = await client.pfs().listFile({
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'putFileFromBytes'}},
      });
      expect(files).toHaveLength(1);
    });
  });

  describe('putFileFromFilePath', () => {
    it('should add a file from a local path to the repo', async () => {
      const client = await createSandbox('putFileFromFilePath');
      const commit = await client.pfs().startCommit({
        branch: {name: 'master', repo: {name: 'putFileFromFilePath'}},
      });

      const initialFiles = await client.pfs().listFile({
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'putFileFromFilePath'}},
      });
      expect(initialFiles).toHaveLength(0);

      await client
        .modifyFile()
        .setCommit(commit)
        .putFileFromFilepath(
          path.join(__dirname, '../../../examples/opencv/images/8MN9Kg0.jpg'),
          '/8MN9Kg0.jpg',
        )
        .end();

      await client.pfs().finishCommit({commit});

      const files = await client.pfs().listFile({
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'putFileFromFilePath'}},
      });
      expect(files).toHaveLength(1);
    });
  });

  describe('deleteFile', () => {
    it('should delete a file by path', async () => {
      const client = await createSandbox('deleteFile');
      const commit = await client.pfs().startCommit({
        branch: {name: 'master', repo: {name: 'deleteFile'}},
      });

      await client
        .modifyFile()
        .setCommit(commit)
        .putFileFromBytes('test.dat', Buffer.from('data'))
        .end();

      await client.pfs().finishCommit({commit});

      const files = await client.pfs().listFile({
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'deleteFile'}},
      });
      expect(files).toHaveLength(1);

      const deleteCommit = await client.pfs().startCommit({
        branch: {name: 'master', repo: {name: 'deleteFile'}},
      });

      await client
        .modifyFile()
        .setCommit(deleteCommit)
        .deleteFile('test.dat')
        .end();

      await client.pfs().finishCommit({commit: deleteCommit});

      const postDeleteFiles = await client.pfs().listFile({
        commitId: deleteCommit.id,
        branch: {name: 'master', repo: {name: 'deleteFile'}},
      });
      expect(postDeleteFiles).toHaveLength(0);
    });
  });
});
