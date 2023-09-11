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

describe('FileSet', () => {
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

      const fileClient = await client.pfs.fileSet();
      const fileSetId = await fileClient
        .putFileFromFilepath(atatPngFilePath, 'at-at.png')
        .putFileFromFilepath(libertyPngFilePath, 'liberty.png')
        .end();

      const commit = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'putFileFromURL'}},
      });

      await client.pfs.addFileSet({projectId: 'default', fileSetId, commit});
      await client.pfs.finishCommit({projectId: 'default', commit});

      const files = await client.pfs.listFile({
        projectId: 'default',
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'putFileFromURL'}},
      });

      expect(files).toHaveLength(2);
    }, 60_000);
  });

  describe('putFileFromBytes', () => {
    it('should add a file in byte format to a repo', async () => {
      const client = await createSandbox('putFileFromBytes');

      const fileClient = await client.pfs.fileSet();
      const fileSetId = await fileClient
        .putFileFromBytes('test.dat', Buffer.from('data'))
        .putFileFromBytes('test2.dat', Buffer.from('data'))
        .end();

      const commit = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'putFileFromBytes'}},
      });
      await client.pfs.addFileSet({projectId: 'default', fileSetId, commit});

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

      const fileClient = await client.pfs.fileSet();

      const fileSetId = await fileClient
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

      const commit = await client.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: 'putFileFromFilePath'}},
      });
      await client.pfs.addFileSet({projectId: 'default', fileSetId, commit});

      await client.pfs.finishCommit({projectId: 'default', commit});

      const files = await client.pfs.listFile({
        projectId: 'default',
        commitId: commit.id,
        branch: {name: 'master', repo: {name: 'putFileFromFilePath'}},
      });
      expect(files).toHaveLength(2);
    });
  });
});
