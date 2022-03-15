import {FileInfo} from '@pachyderm/node-pachyderm';

import {MutationResolvers, QueryResolvers} from '@dash-backend/generated/types';
import {FILE_DOWNLOAD_LIMIT} from '@dash-backend/lib/constants';
import formatBytes from '@dash-backend/lib/formatBytes';
import formatDiff from '@dash-backend/lib/formatDiff';
import {toGQLFileType} from '@dash-backend/lib/gqlEnumMappers';

interface FileResolver {
  Query: {
    files: QueryResolvers['files'];
  };
  Mutation: {
    putFilesFromURLs: MutationResolvers['putFilesFromURLs'];
    deleteFile: MutationResolvers['deleteFile'];
  };
}

const getDownloadLink = (file: FileInfo.AsObject, host: string) => {
  const repoName = file.file?.commit?.branch?.repo?.name;
  const branchName = file.file?.commit?.branch?.name;
  const commitId = file.file?.commit?.id;
  const filePath = file.file?.path;

  if (repoName && branchName && commitId && filePath) {
    return `${
      process.env.NODE_ENV === 'test' ? 'http:' : ''
    }${host}/download/${repoName}/${branchName}/${commitId}${filePath}`;
  }

  return null;
};

const fileResolver: FileResolver = {
  Query: {
    files: async (
      _parent,
      {args: {commitId, path, branchName, repoName}},
      {pachClient, host},
    ) => {
      const files = await pachClient.pfs().listFile({
        commitId: commitId || 'master',
        path: path || '/',
        branch: {name: branchName, repo: {name: repoName}},
      });

      const diff = await pachClient.pfs().diffFile({
        commitId: commitId || 'master',
        path: path || '/',
        branch: {name: branchName, repo: {name: repoName}},
      });

      const formattedDiff = formatDiff(diff);

      return {
        diff: formattedDiff.diff,
        files: files.map((file) => ({
          commitId: file.file?.commit?.id || '',
          committed: file.committed,
          // TODO: This may eventually come from the S3 gateway or Pach's http server
          download: getDownloadLink(file, host),
          downloadDisabled: (file.sizeBytes || 0) > FILE_DOWNLOAD_LIMIT,
          hash: file.hash.toString() || '',
          path: file.file?.path || '/',
          repoName: file.file?.commit?.branch?.repo?.name || '',
          sizeBytes: file.sizeBytes || 0,
          sizeDisplay: formatBytes(file.sizeBytes || 0),
          type: toGQLFileType(file.fileType),
          commitAction: formattedDiff.diffTotals[file.file?.path || ''],
        })),
      };
    },
  },
  Mutation: {
    putFilesFromURLs: async (
      _field,
      {args: {branch, files, repo}},
      {pachClient},
    ) => {
      const fileClient = await pachClient.pfs().modifyFile();
      fileClient.autoCommit({name: branch, repo: {name: repo}});
      files.forEach((file) => {
        fileClient.putFileFromURL(file.path, file.url);
      });
      await fileClient.end();
      return files.map((file) => file.url);
    },
    deleteFile: async (
      _field,
      {args: {repo, branch, filePath, force}},
      {pachClient},
    ) => {
      const fileClient = await pachClient.pfs().fileSet();

      const fileSetId = await fileClient.deleteFile(filePath).end();

      const commit = await pachClient.pfs().startCommit({
        branch: {name: branch, repo: {name: repo}},
      });
      await pachClient.pfs().addFileSet({fileSetId, commit});
      await pachClient.pfs().finishCommit({commit});
      return commit.id;
    },
  },
};

export default fileResolver;
