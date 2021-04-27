import {FileInfo} from '@pachyderm/proto/pb/pfs/pfs_pb';

import {QueryResolvers} from '@dash-backend/generated/types';
import {toGQLFileType} from '@dash-backend/lib/gqlEnumMappers';

interface FileResolver {
  Query: {
    files: QueryResolvers['files'];
  };
}

const getDownloadLink = (file: FileInfo.AsObject) => {
  if (
    file.file?.commit?.repo?.name &&
    file.file?.commit?.id &&
    file.file?.path
  ) {
    return `/download/${file.file?.commit?.repo?.name}/${file.file?.commit?.id}${file?.file?.path}`;
  }

  return null;
};

const fileResolver: FileResolver = {
  Query: {
    files: async (
      _parent,
      {args: {commitId, path, repoName}},
      {pachClient},
    ) => {
      const files = await pachClient.pfs().listFile({
        commitId: commitId || 'master',
        path: path || '/',
        repoName,
      });

      return files.map((file) => ({
        commitId: file.file?.commit?.id || '',
        committed: file.committed,
        // TODO: This may eventually come from the S3 gateway or Pach's http server
        download: getDownloadLink(file),
        hash: file.hash.toString(),
        path: file.file?.path || '/',
        repoName: file.file?.commit?.repo?.name || '',
        sizeBytes: file.sizeBytes,
        type: toGQLFileType(file.fileType),
      }));
    },
  },
};

export default fileResolver;
