import {FileInfo} from '@pachyderm/proto/pb/pfs/pfs_pb';

import {QueryResolvers} from '@dash-backend/generated/types';
import {toGQLFileType} from '@dash-backend/lib/gqlEnumMappers';

interface FileResolver {
  Query: {
    files: QueryResolvers['files'];
  };
}

const getDownloadLink = (file: FileInfo.AsObject) => {
  const repoName = file.file?.commit?.branch?.repo?.name;
  const branchName = file.file?.commit?.branch?.name;
  const commitId = file.file?.commit?.id;
  const filePath = file.file?.path;

  if (repoName && branchName && commitId && filePath) {
    return `/download/${repoName}/${branchName}/${commitId}${filePath}`;
  }

  return null;
};

const fileResolver: FileResolver = {
  Query: {
    files: async (
      _parent,
      {args: {commitId, path, branchName, repoName}},
      {pachClient},
    ) => {
      const files = await pachClient.pfs().listFile({
        commitId: commitId || 'master',
        path: path || '/',
        branch: {name: branchName, repo: {name: repoName}},
      });

      return files.map((file) => ({
        commitId: file.file?.commit?.id || '',
        committed: file.committed,
        // TODO: This may eventually come from the S3 gateway or Pach's http server
        download: getDownloadLink(file),
        hash: file.hash.toString(),
        path: file.file?.path || '/',
        repoName: file.file?.commit?.branch?.repo?.name || '',
        sizeBytes: file.sizeBytes,
        type: toGQLFileType(file.fileType),
      }));
    },
  },
};

export default fileResolver;
