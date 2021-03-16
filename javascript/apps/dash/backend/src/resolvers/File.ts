import {QueryResolvers} from '@dash-backend/generated/types';
import client from '@dash-backend/grpc/client';

interface FileResolver {
  Query: {
    files: QueryResolvers['files'];
  };
}

const fileResolver: FileResolver = {
  Query: {
    files: async (
      _parent,
      {args: {commitId, path, repoName}},
      {pachdAddress = '', authToken = '', log},
    ) => {
      const files = await client({pachdAddress, authToken, log})
        .pfs()
        .listFile({
          commitId: commitId || 'master',
          path: path || '/',
          repoName,
        });

      return files.map((file) => ({
        commitId: file.file?.commit?.id || '',
        // TODO: This may eventually come from the S3 gateway or Pach's http server
        download: `/download/${file.file?.commit?.repo?.name}/${file.file?.commit?.id}/${file?.file?.path}`,
        hash: file.hash.toString(),
        path: file.file?.path || '/',
        repoName: file.file?.commit?.repo?.name || '',
        sizeBytes: file.sizeBytes,
        type: file.fileType,
      }));
    },
  },
};

export default fileResolver;
