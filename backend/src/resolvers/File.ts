import {DEFAULT_FILE_LIMIT} from '@dash-backend/constants/limits';
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
import {FileInfo, CommitState} from '@dash-backend/proto';

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
  const projectId = file.file?.commit?.branch?.repo?.project?.name;

  if (repoName && branchName && commitId && filePath) {
    return `${
      process.env.NODE_ENV === 'test' ? 'http:' : ''
    }${host}/download/${projectId}/${repoName}/${branchName}/${commitId}${filePath}`;
  }

  return null;
};

const fileResolver: FileResolver = {
  Query: {
    files: async (
      _parent,
      {
        args: {
          projectId,
          commitId,
          path,
          branchName,
          repoName,
          cursorPath,
          limit,
          reverse,
        },
      },
      {pachClient, host},
    ) => {
      let diff:
        | undefined
        | {diffTotals: Record<string, FileCommitState>; diff: Diff};
      limit = limit || DEFAULT_FILE_LIMIT;

      const files = await pachClient.pfs().listFile({
        projectId,
        commitId: commitId || '',
        path: path || '/',
        branch: {
          name: branchName,
          repo: {name: repoName},
        },
        cursorPath: cursorPath || '',
        number: limit + 1,
        reverse: reverse || undefined,
      });

      const commit = commitInfoToGQLCommit(
        await pachClient.pfs().inspectCommit({
          projectId,
          wait: CommitState.COMMIT_STATE_UNKNOWN,
          commit: {
            id: commitId || '',
            branch: {
              name: branchName,
              repo: {name: repoName},
            },
          },
        }),
      );

      if (commit.originKind !== OriginKind.ALIAS && commit.finished !== -1) {
        const diffResponse = await pachClient.pfs().diffFile({
          projectId,
          newFileObject: {
            commitId: commitId || '',
            path: path || '/',
            branch: {
              name: branchName,
              repo: {name: repoName},
            },
          },
        });
        diff = formatDiff(diffResponse);
      }

      let nextCursor = undefined;

      //If files.length is not greater than limit there are no pages left
      if (files.length > limit) {
        // cursor file is included in next page
        nextCursor = files[files.length - 1].file?.path;
        //remove the extra file from the response
        files.pop();
      }

      return {
        diff: diff?.diff,
        cursor: nextCursor,
        hasNextPage: !!nextCursor,
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
      {args: {projectId, repo, branch, filePath}},
      {pachClient},
    ) => {
      const fileClient = await pachClient.pfs().fileSet();

      const fileSetId = await fileClient.deleteFile(filePath).end();

      const commit = await pachClient.pfs().startCommit({
        projectId,
        branch: {name: branch, repo: {name: repo}},
      });
      await pachClient.pfs().addFileSet({projectId, fileSetId, commit});
      await pachClient.pfs().finishCommit({projectId, commit});
      return commit.id;
    },
  },
};

export default fileResolver;
