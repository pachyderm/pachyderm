import {APIClient} from '@pachyderm/proto/pb/pfs/pfs_grpc_pb';
import {
  FileInfo,
  GetFileRequest,
  ListFileRequest,
  ListRepoRequest,
  RepoInfo,
} from '@pachyderm/proto/pb/pfs/pfs_pb';
import {BytesValue} from 'google-protobuf/google/protobuf/wrappers_pb';
import flattenDeep from 'lodash/flattenDeep';

import {
  fileFromObject,
  fileInfoFromObject,
  FileObject,
} from '@dash-backend/grpc/builders';
import {ServiceArgs} from '@dash-backend/lib/types';

const pfs = ({
  pachdAddress,
  channelCredentials,
  credentialMetadata,
  log,
}: ServiceArgs) => {
  const client = new APIClient(pachdAddress, channelCredentials);

  return {
    listFile: (params: FileObject) => {
      const listFileRequest = new ListFileRequest();
      const file = fileFromObject(params);

      listFileRequest.setFile(file);

      const stream = client.listFile(listFileRequest, credentialMetadata);

      return new Promise<FileInfo.AsObject[]>((resolve, reject) => {
        const files: FileInfo.AsObject[] = [];

        stream.on('data', (chunk) => {
          const [
            repoName,
            commitId,
            path,
            fileType,
            sizeBytes,
            ,
            ,
            ,
            hash,
          ] = flattenDeep(chunk.array);

          files.push(
            fileInfoFromObject({
              file: {commitId, path, repoName},
              fileType,
              sizeBytes,
              hash,
            }).toObject(),
          );
        });
        stream.on('error', (err) => reject(err));
        stream.on('end', () => resolve(files));
      });
    },
    getFile: (params: FileObject) => {
      const getFileRequest = new GetFileRequest();
      const file = fileFromObject(params);

      getFileRequest.setFile(file);

      const stream = client.getFile(getFileRequest, credentialMetadata);

      return new Promise<Buffer>((resolve, reject) => {
        // The chunks contain:
        // chunks[0]: File metadata
        // chunks[1]: File data
        const chunks: BytesValue[] = [];

        stream.on('data', (chunk) => {
          chunks.push(chunk.array[0]);
        });

        stream.on('end', () => {
          // We typically already have the metadata we need, so we create a buffer from chunks[1]
          if (chunks[1]) {
            return resolve(Buffer.from(chunks[1]));
          } else {
            return reject(new Error('File does not exist.'));
          }
        });

        stream.on('error', (err) => {
          return reject(err);
        });
      });
    },
    listRepo: () => {
      return new Promise<RepoInfo.AsObject[]>((resolve, reject) => {
        log.info('listRepo request');

        client.listRepo(
          new ListRepoRequest(),
          credentialMetadata,
          (error, res) => {
            if (error) {
              log.error({error: error.message}, 'listRepo request failed');
              return reject(error);
            }

            log.error('listRepo request succeeded');
            return resolve(res.toObject().repoInfoList);
          },
        );
      });
    },
  };
};

export default pfs;
