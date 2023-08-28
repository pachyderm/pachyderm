import {TextEncoder} from 'util';

import {ApolloError} from 'apollo-server-errors';

import {DEFAULT_FILE_LIMIT} from '@dash-backend/constants/limits';
import {
  MutationResolvers,
  QueryResolvers,
  FileCommitState,
} from '@dash-backend/generated/types';
import {appendBuffer} from '@dash-backend/lib/appendBuffer';
import {
  FILE_DOWNLOAD_API_VERSION,
  FILE_DOWNLOAD_LIMIT,
} from '@dash-backend/lib/constants';
import formatBytes from '@dash-backend/lib/formatBytes';
import {formatDiffOnlyTotals} from '@dash-backend/lib/formatDiff';
import {toGQLFileType} from '@dash-backend/lib/gqlEnumMappers';
import {zstdCompress} from '@dash-backend/lib/zstd';
import {FileInfo, CommitState, FileType} from '@dash-backend/proto';

import {commitInfoToGQLCommit} from './builders/pfs';
interface FileResolver {
  Query: {
    files: QueryResolvers['files'];
    fileDownload: QueryResolvers['fileDownload'];
  };
  Mutation: {
    putFilesFromURLs: MutationResolvers['putFilesFromURLs'];
    deleteFiles: MutationResolvers['deleteFiles'];
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
      let diffTotals: Record<string, FileCommitState>;
      limit = limit || DEFAULT_FILE_LIMIT;

      const files = await pachClient.pfs.listFile({
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
        await pachClient.pfs.inspectCommit({
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

      if (commit.finished !== -1) {
        const diffResponse = await pachClient.pfs.diffFile({
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
        diffTotals = formatDiffOnlyTotals(diffResponse);
      }

      let nextCursor = undefined;

      //If files.length is not greater than limit there are no pages left
      if (files.length > limit) {
        files.pop(); //remove the extra file from the response
        nextCursor = files[files.length - 1].file?.path;
      }

      return {
        cursor: nextCursor,
        hasNextPage: !!nextCursor,
        files: files.map((file) => ({
          commitId: file.file?.commit?.id || '',
          committed: file.committed,
          download:
            file.fileType !== FileType.DIR &&
            file.sizeBytes <= FILE_DOWNLOAD_LIMIT
              ? getDownloadLink(file, host)
              : undefined,
          hash: file.hash.toString() || '',
          path: file.file?.path || '/',
          repoName: file.file?.commit?.branch?.repo?.name || '',
          sizeBytes: file.sizeBytes || 0,
          sizeDisplay: formatBytes(file.sizeBytes || 0),
          type: toGQLFileType(file.fileType),
          commitAction: diffTotals
            ? diffTotals[file.file?.path || '']
            : undefined,
        })),
      };
    },
    // This resolver follows steps laid out here to generate a url that pach can consume
    // https://www.notion.so/2023-04-03-HTTP-file-and-archive-downloads-cfb56fac16e54957b015070416b09e94
    fileDownload: async (
      _parent,
      {args: {projectId, repoId, commitId, paths}},
      {log},
    ) => {
      try {
        // 1. Sort the list of paths asciibetically.
        const sortedPaths = paths
          .map((p) => `${projectId}/${repoId}@${commitId}:${p}`)
          .sort();

        // 2. Terminate each file name with a `0x00` byte and concatenate to one string. A request MUST end with a `0x00` byte.
        const terminatedString = sortedPaths.join('\x00');
        // 2. a. encode to Uint8Array
        const encoder = new TextEncoder();
        const encodedData = encoder.encode(terminatedString);

        // 3. Compress the result with [Zstandard](https://facebook.github.io/zstd/).
        const compressedData = await zstdCompress(encodedData);

        // 4. Prefix the result with `0x01` (for URL format version 1).
        const prefixedData = appendBuffer(
          new Uint8Array([FILE_DOWNLOAD_API_VERSION]),
          compressedData,
        );

        // 5. Convert the result to URL-safe Base64.  See section 3.5 of RFC4648 for a description of this coding.  Padding may be omitted.
        const base64url = Buffer.from(prefixedData).toString('base64url');

        // 6. Prefix the result with `/archive` and suffix it with the desired archive format, like `.tar.zst`
        return `/archive/${base64url}.zip`;
      } catch (error) {
        log.error({err: error}, 'Error generating archive URL');
        throw new ApolloError('Error generating archive URL');
      }
    },
  },
  Mutation: {
    putFilesFromURLs: async (
      _field,
      {args: {branch, files, repo}},
      {pachClient},
    ) => {
      const fileClient = await pachClient.pfs.modifyFile();
      fileClient.autoCommit({name: branch, repo: {name: repo}});
      files.forEach((file) => {
        fileClient.putFileFromURL(file.path, file.url);
      });
      await fileClient.end();
      return files.map((file) => file.url);
    },
    deleteFiles: async (
      _field,
      {args: {projectId, repo, branch, filePaths}},
      {pachClient, log},
    ) => {
      const deleteCommit = await pachClient.pfs.startCommit({
        projectId,
        branch: {name: branch, repo: {name: repo}},
      });

      try {
        const modifyFileClient = await pachClient.pfs.modifyFile();
        await modifyFileClient
          .setCommit(deleteCommit)
          .deleteFiles(filePaths)
          .end();
      } catch (error) {
        log.error({err: error}, 'Error deleting files');
        throw new ApolloError('Error deleting files');
      }
      await pachClient.pfs.finishCommit({projectId, commit: deleteCommit});
      return deleteCommit.id;
    },
  },
};

export default fileResolver;
