import path from 'path';

import client from '../../../../client';

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

      const fileClient = await client.pfs().modifyFile();
      await fileClient
        .setCommit(commit)
        .putFileFromURL('at-at.png', 'http://imgur.com/8MN9Kg0.png')
        .putFileFromURL('liberty.png', 'http://imgur.com/46Q8nDz.png')
        .end();

      await client.pfs().finishCommit({commit});

      const files = await client.pfs().listFile({
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'putFileFromURL'}},
      });

      expect(files).toHaveLength(2);
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

      const fileClient = await client.pfs().modifyFile();
      await fileClient
        .setCommit(commit)
        .putFileFromBytes('test.dat', Buffer.from('data'))
        .putFileFromBytes('test2.dat', Buffer.from('data'))
        .end();
      await client.pfs().finishCommit({commit});

      const files = await client.pfs().listFile({
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'putFileFromBytes'}},
      });
      expect(files).toHaveLength(2);
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

      const fileClient = await client.pfs().modifyFile();

      await fileClient
        .setCommit(commit)
        .putFileFromFilepath(
          path.join(
            __dirname,
            '../../../../../examples/opencv/images/8MN9Kg0.jpg',
          ),
          '/8MN9Kg0.jpg',
        )
        .putFileFromFilepath(
          path.join(
            __dirname,
            '../../../../../examples/opencv/images/46Q8nDz.jpg',
          ),
          '/46Q8nDz.jpg',
        )
        .end();

      await client.pfs().finishCommit({commit});

      const files = await client.pfs().listFile({
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'putFileFromFilePath'}},
      });
      expect(files).toHaveLength(2);
    });
  });

  describe('deleteFile', () => {
    it('should delete a file by path', async () => {
      const client = await createSandbox('deleteFile');
      const commit = await client.pfs().startCommit({
        branch: {name: 'master', repo: {name: 'deleteFile'}},
      });

      const fileClient = await client.pfs().modifyFile();

      await fileClient
        .setCommit(commit)
        .putFileFromBytes('test.dat', Buffer.from('data'))
        .putFileFromBytes('test2.dat', Buffer.from('data'))
        .end();

      await client.pfs().finishCommit({commit});

      const files = await client.pfs().listFile({
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'deleteFile'}},
      });
      expect(files).toHaveLength(2);

      const deleteCommit = await client.pfs().startCommit({
        branch: {name: 'master', repo: {name: 'deleteFile'}},
      });

      const fileClient2 = await client.pfs().modifyFile();

      await fileClient2
        .setCommit(deleteCommit)
        .deleteFile('test.dat')
        .deleteFile('test2.dat')
        .end();

      await client.pfs().finishCommit({commit: deleteCommit});

      const postDeleteFiles = await client.pfs().listFile({
        commitId: deleteCommit.id,
        branch: {name: 'master', repo: {name: 'deleteFile'}},
      });
      expect(postDeleteFiles).toHaveLength(0);
    });
  });

  describe('autoCommit', () => {
    it('should be able to use auto commits', async () => {
      const client = await createSandbox('putFileFromFilePath');

      const commits = await client
        .pfs()
        .listCommit({repo: {name: 'putFileFromFilePath'}});
      expect(commits).toHaveLength(0);

      const fileClient = await client.pfs().modifyFile();

      await fileClient
        .autoCommit({name: 'master', repo: {name: 'putFileFromFilePath'}})
        .putFileFromFilepath(
          path.join(
            __dirname,
            '../../../../../examples/opencv/images/8MN9Kg0.jpg',
          ),
          '/8MN9Kg0.jpg',
        )
        .putFileFromFilepath(
          path.join(
            __dirname,
            '../../../../../examples/opencv/images/46Q8nDz.jpg',
          ),
          '/46Q8nDz.jpg',
        )
        .end();

      const newCommits = await client
        .pfs()
        .listCommit({repo: {name: 'putFileFromFilePath'}});
      expect(newCommits).toHaveLength(1);
      const files = await client.pfs().listFile({
        commitId: '',
        branch: {name: 'master', repo: {name: 'putFileFromFilePath'}},
      });
      expect(files).toHaveLength(2);
    });
  });
});
