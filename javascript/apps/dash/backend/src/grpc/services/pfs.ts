import {APIClient} from '@pachyderm/proto/pb/pfs/pfs_grpc_pb';
import {
  CommitInfo,
  FileInfo,
  GetFileRequest,
  InspectRepoRequest,
  ListCommitRequest,
  ListFileRequest,
  ListRepoRequest,
  RepoInfo,
} from '@pachyderm/proto/pb/pfs/pfs_pb';
import {BytesValue} from 'google-protobuf/google/protobuf/wrappers_pb';

import {
  fileFromObject,
  FileObject,
  repoFromObject,
  RepoObject,
} from '@dash-backend/grpc/builders/pfs';
import {ServiceArgs} from '@dash-backend/lib/types';

const pfs = ({
  pachdAddress,
  channelCredentials,
  credentialMetadata,
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

        stream.on('data', (chunk: FileInfo) => files.push(chunk.toObject()));
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
        // chunks[2...n]: All 0s
        // TODO: why is 2...n even sent?
        const chunks: BytesValue.AsObject['value'][] = [];

        stream.on('data', (chunk: BytesValue) => {
          chunks.push(chunk.getValue());
        });

        stream.on('end', () => {
          // We already have the metadata we need, so we only create a buffer from chunks[1][0]
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
    listCommit: (
      repoName: RepoObject['name'],
      limit?: number,
      reverse = true,
    ) => {
      const listCommitRequest = new ListCommitRequest();
      const repo = repoFromObject({name: repoName});

      listCommitRequest.setRepo(repo);
      listCommitRequest.setReverse(reverse);

      if (limit) {
        listCommitRequest.setNumber(limit);
      }

      const stream = client.listCommit(listCommitRequest, credentialMetadata);

      return new Promise<CommitInfo.AsObject[]>((resolve, reject) => {
        const commits: CommitInfo.AsObject[] = [];

        stream.on('data', (chunk: CommitInfo) =>
          commits.push(chunk.toObject()),
        );
        stream.on('error', (err) => reject(err));
        stream.on('end', () => resolve(commits));
      });
    },
    listRepo: () => {
      return new Promise<RepoInfo.AsObject[]>((resolve, reject) => {
        client.listRepo(
          new ListRepoRequest(),
          credentialMetadata,
          (error, res) => {
            if (error) {
              return reject(error);
            }

            return resolve(res.toObject().repoInfoList);
          },
        );
      });
    },
    inspectRepo: (name: RepoObject['name']) => {
      return new Promise<RepoInfo.AsObject>((resolve, reject) => {
        const inspectRepoRequest = new InspectRepoRequest();
        const repo = repoFromObject({name});

        inspectRepoRequest.setRepo(repo);

        client.inspectRepo(
          inspectRepoRequest,
          credentialMetadata,
          (error, res) => {
            if (error) {
              return reject(error);
            }

            return resolve(res.toObject());
          },
        );
      });
    },
  };
};

export default pfs;
