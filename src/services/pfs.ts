import {APIClient} from '@pachyderm/proto/pb/pfs/pfs_grpc_pb';
import {
  BranchInfo,
  CommitInfo,
  CommitSetInfo,
  FileInfo,
  GetFileRequest,
  InspectFileRequest,
  InspectRepoRequest,
  InspectBranchRequest,
  ListCommitSetRequest,
  ListFileRequest,
  ListRepoRequest,
  RepoInfo,
  SquashCommitSetRequest,
  Commit,
  ClearCommitRequest,
} from '@pachyderm/proto/pb/pfs/pfs_pb';
import {Empty} from 'google-protobuf/google/protobuf/empty_pb';
import {BytesValue} from 'google-protobuf/google/protobuf/wrappers_pb';
import {extract} from 'tar-stream';

import {
  commitSetFromObject,
  CommitSetObject,
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
  inspectCommitSetRequestFromObject,
  InspectCommitSetRequestObject,
  repoFromObject,
  RepoObject,
  BranchObject,
  StartCommitRequestObject,
  CommitObject,
  startCommitRequestFromObject,
  FinishCommitRequestObject,
  finishCommitRequestFromObject,
  InspectCommitRequestObject,
  inspectCommitRequestFromObject,
  commitFromObject,
  ListCommitRequestObject,
  listCommitRequestFromObject,
  SubscribeCommitRequestObject,
  subscribeCommitRequestFromObject,
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
    inspectFile: (params: FileObject) => {
      return new Promise<FileInfo.AsObject>((resolve, reject) => {
        const inspectFileRequest = new InspectFileRequest();
        const file = fileFromObject(params);

        inspectFileRequest.setFile(file);

        client.inspectFile(
          inspectFileRequest,
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
    listCommit: (request: ListCommitRequestObject) => {
      const listCommitRequest = listCommitRequestFromObject(request);
      const stream = client.listCommit(listCommitRequest, credentialMetadata);

      return streamToObjectArray<CommitInfo, CommitInfo.AsObject>(stream);
    },
    startCommit: (request: StartCommitRequestObject) => {
      return new Promise<Commit.AsObject>((resolve, reject) => {
        const startCommitRequest = startCommitRequestFromObject(request);

        client.startCommit(
          startCommitRequest,
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
    finishCommit: (request: FinishCommitRequestObject) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const finishCommitRequest = finishCommitRequestFromObject(request);

        client.finishCommit(
          finishCommitRequest,
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
    clearCommit: (params: CommitObject) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const clearCommitRequest = new ClearCommitRequest().setCommit(
          commitFromObject(params),
        );

        client.clearCommit(clearCommitRequest, credentialMetadata, (error) => {
          if (error) {
            return reject(error);
          }
          return resolve({});
        });
      });
    },
    inspectCommit: (request: InspectCommitRequestObject) => {
      return new Promise<CommitInfo.AsObject>((resolve, reject) => {
        const inspectCommitRequest = inspectCommitRequestFromObject(request);

        client.inspectCommit(
          inspectCommitRequest,
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
    subscribeCommit: (request: SubscribeCommitRequestObject) => {
      const subscribeCommitRequest = subscribeCommitRequestFromObject(request);
      const stream = client.subscribeCommit(
        subscribeCommitRequest,
        credentialMetadata,
      );

      return streamToObjectArray<CommitInfo, CommitInfo.AsObject>(stream);
    },
    inspectCommitSet: (request: InspectCommitSetRequestObject) => {
      const inspectCommitSetRequest =
        inspectCommitSetRequestFromObject(request);
      const stream = client.inspectCommitSet(
        inspectCommitSetRequest,
        credentialMetadata,
      );

      return streamToObjectArray<CommitInfo, CommitInfo.AsObject>(stream);
    },
    listCommitSet: () => {
      const listCommitSetRequest = new ListCommitSetRequest();
      const stream = client.listCommitSet(
        listCommitSetRequest,
        credentialMetadata,
      );

      return streamToObjectArray<CommitSetInfo, CommitSetInfo.AsObject>(stream);
    },
    squashCommitSet: (commitSet: CommitSetObject) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const squashCommitSetRequest =
          new SquashCommitSetRequest().setCommitSet(
            commitSetFromObject(commitSet),
          );

        client.squashCommitSet(
          squashCommitSetRequest,
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
    deleteAll: () => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        client.deleteAll(new Empty(), (error) => {
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
