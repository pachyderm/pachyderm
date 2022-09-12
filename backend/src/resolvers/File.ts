import {FileInfo, CommitState} from '@pachyderm/node-pachyderm';

import {
  MutationResolvers,
  QueryResolvers,
  FileCommitState,
  OriginKind,
  Diff,
} from '@dash-backend/generated/types';
import {FILE_DOWNLOAD_LIMIT} from '@dash-backend/lib/constants';
import formatBytes from '@dash-backend/lib/formatBytes';
import formatDiff from '@dash-backend/lib/formatDiff';
import {toGQLFileType} from '@dash-backend/lib/gqlEnumMappers';

import {commitInfoToGQLCommit} from './builders/pfs';

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
      let diff:
        | undefined
        | {diffTotals: Record<string, FileCommitState>; diff: Diff};

      const files = await pachClient.pfs().listFile({
        commitId: commitId || 'master',
        path: path || '/',
        branch: {name: branchName, repo: {name: repoName}},
      });

      const commit = commitInfoToGQLCommit(
        await pachClient.pfs().inspectCommit({
          wait: CommitState.COMMIT_STATE_UNKNOWN,
          commit: {
            id: commitId || '',
            branch: {name: branchName || 'master', repo: {name: repoName}},
          },
        }),
      );

      if (commit.originKind !== OriginKind.ALIAS && commit.finished !== -1) {
        const diffResponse = await pachClient.pfs().diffFile({
          commitId: commitId || 'master',
          path: path || '/',
          branch: {name: branchName, repo: {name: repoName}},
        });
        diff = formatDiff(diffResponse);
      }

      return {
        diff: diff?.diff,
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
          commitAction: diff
            ? diff.diffTotals[file.file?.path || '']
            : undefined,
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
      {args: {repo, branch, filePath}},
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
