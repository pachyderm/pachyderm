import {APIClient} from '@pachyderm/proto/pb/pfs/pfs_grpc_pb';
import {
  BranchInfo,
  CommitInfo,
  FileInfo,
  GetFileRequest,
  InspectRepoRequest,
  InspectBranchRequest,
  ListCommitRequest,
  ListFileRequest,
  ListRepoRequest,
  RepoInfo,
} from '@pachyderm/proto/pb/pfs/pfs_pb';
import {Empty} from 'google-protobuf/google/protobuf/empty_pb';
import {BytesValue} from 'google-protobuf/google/protobuf/wrappers_pb';
import {extract} from 'tar-stream';

import {
  createBranchRequestFromObject,
  CreateBranchRequestObject,
  listBranchRequestFromObject,
  ListBranchRequestObject,
  deleteBranchRequestFromObject,
  DeleteBranchRequestObject,
  createRepoRequestFromObject,
  CreateRepoRequestObject,
  deleteRepoRequestFromObject,
  DeleteRepoRequestObject,
  fileFromObject,
  branchFromObject,
  FileObject,
  repoFromObject,
  RepoObject,
  BranchObject,
} from '../builders/pfs';
import {ServiceArgs} from '../lib/types';
import streamToObjectArray from '../utils/streamToObjectArray';

import {GRPC_MAX_MESSAGE_LENGTH} from './constants/pfs';

const pfs = ({
  pachdAddress,
  channelCredentials,
  credentialMetadata,
}: ServiceArgs) => {
  const client = new APIClient(pachdAddress, channelCredentials, {
    /* eslint-disable @typescript-eslint/naming-convention */
    'grpc.max_receive_message_length': GRPC_MAX_MESSAGE_LENGTH,
    'grpc.max_send_message_length': GRPC_MAX_MESSAGE_LENGTH,
    /* eslint-enable @typescript-eslint/naming-convention */
  });

  return {
    listFile: (params: FileObject) => {
      const listFileRequest = new ListFileRequest();
      const file = fileFromObject(params);

      listFileRequest.setFile(file);
      listFileRequest.setDetails(true);

      const stream = client.listFile(listFileRequest, credentialMetadata);

      return streamToObjectArray<FileInfo, FileInfo.AsObject>(stream);
    },
    getFile: (params: FileObject) => {
      const getFileRequest = new GetFileRequest();
      const file = fileFromObject(params);

      getFileRequest.setFile(file);

      const stream = client.getFileTAR(getFileRequest, credentialMetadata);

      return new Promise<Buffer>((resolve, reject) => {
        // The GetFile request returns a tar stream.
        // We have to untar it before we can read it.
        const buffers: Buffer[] = [];
        const extractor = extract();

        extractor.on('entry', (_, estream, next) => {
          estream.on('data', (buffer: Buffer) => buffers.push(buffer));
          estream.on('end', () => next());
          estream.resume();
        });

        extractor.on('finish', () => {
          if (buffers.length) {
            return resolve(Buffer.concat(buffers));
          } else {
            return reject(new Error('File does not exist.'));
          }
        });

        stream.on('data', (chunk: BytesValue) =>
          extractor.write(chunk.getValue()),
        );
        stream.on('end', () => extractor.end());
        stream.on('error', (err) => reject(err));
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

      return streamToObjectArray<CommitInfo, CommitInfo.AsObject>(stream);
    },
    createBranch: (request: CreateBranchRequestObject) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const createBranchRequest = createBranchRequestFromObject(request);

        client.createBranch(
          createBranchRequest,
          credentialMetadata,
          (error) => {
            if (error) {
              return reject(error);
            }
            return resolve({});
          },
        );
      });
    },
    inspectBranch: (params: BranchObject) => {
      return new Promise<BranchInfo.AsObject>((resolve, reject) => {
        const inspectBranchRequest = new InspectBranchRequest();
        const branch = branchFromObject(params);

        inspectBranchRequest.setBranch(branch);

        client.inspectBranch(
          inspectBranchRequest,
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
    listBranch: (request: ListBranchRequestObject) => {
      const listBranchRequest = listBranchRequestFromObject(request);
      const stream = client.listBranch(listBranchRequest, credentialMetadata);

      return streamToObjectArray<BranchInfo, BranchInfo.AsObject>(stream);
    },
    deleteBranch: (request: DeleteBranchRequestObject) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const deleteBranchRequest = deleteBranchRequestFromObject(request);

        client.deleteBranch(
          deleteBranchRequest,
          credentialMetadata,
          (error) => {
            if (error) {
              return reject(error);
            }
            return resolve({});
          },
        );
      });
    },
    listRepo: (type = 'user') => {
      const listRepoRequest = new ListRepoRequest().setType(type);
      const stream = client.listRepo(listRepoRequest, credentialMetadata);

      return streamToObjectArray<RepoInfo, RepoInfo.AsObject>(stream);
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
    createRepo: (request: CreateRepoRequestObject) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const createRepoRequest = createRepoRequestFromObject(request);
        client.createRepo(createRepoRequest, credentialMetadata, (error) => {
          if (error) {
            return reject(error);
          }
          return resolve({});
        });
      });
    },
    deleteRepo: (request: DeleteRepoRequestObject) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const deleteRepoRequest = deleteRepoRequestFromObject(request);
        client.deleteRepo(deleteRepoRequest, credentialMetadata, (error) => {
          if (error) {
            return reject(error);
          }
          return resolve({});
        });
      });
    },
  };
};

export default pfs;
