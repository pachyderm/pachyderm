import {Empty} from 'google-protobuf/google/protobuf/empty_pb';
import {BytesValue} from 'google-protobuf/google/protobuf/wrappers_pb';

import {
  commitSetFromObject,
  CommitSetObject,
  fileFromObject,
  branchFromObject,
  FileObject,
  repoFromObject,
  RepoObject,
  BranchObject,
  CommitObject,
  commitFromObject,
  triggerFromObject,
} from '../../builders/pfs';
import {
  ServiceArgs,
  CreateBranchArgs,
  CreateRepoRequestArgs,
  DeleteBranchRequestArgs,
  DeleteRepoRequestArgs,
  FinishCommitRequestArgs,
  InspectCommitRequestArgs,
  InspectCommitSetArgs,
  ListBranchRequestArgs,
  ListCommitArgs,
  StartCommitRequestArgs,
  SubscribeCommitRequestArgs,
  RenewFileSetRequestArgs,
  AddFileSetRequestArgs,
} from '../../lib/types';
import {APIClient} from '../../proto/pfs/pfs_grpc_pb';
import {
  BranchInfo,
  CommitInfo,
  CommitSetInfo,
  FileInfo,
  DiffFileResponse,
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
  ListCommitRequest,
  Branch,
  CreateBranchRequest,
  DropCommitSetRequest,
  InspectCommitSetRequest,
  StartCommitRequest,
  FinishCommitRequest,
  CreateRepoRequest,
  InspectCommitRequest,
  SubscribeCommitRequest,
  Repo,
  ListBranchRequest,
  DeleteBranchRequest,
  DeleteRepoRequest,
  AddFileSetRequest,
  RenewFileSetRequest,
  DiffFileRequest,
  ComposeFileSetRequest,
} from '../../proto/pfs/pfs_pb';
import streamToObjectArray from '../../utils/streamToObjectArray';
import {RPC_DEADLINE_MS} from '../constants/rpc';

import {FileSet} from './clients/FileSet';
import {ModifyFile} from './clients/ModifyFile';
import {GRPC_MAX_MESSAGE_LENGTH} from './lib/constants';

const pfs = ({
  pachdAddress,
  channelCredentials,
  credentialMetadata,
  plugins = [],
}: ServiceArgs) => {
  const client = new APIClient(pachdAddress, channelCredentials, {
    /* eslint-disable @typescript-eslint/naming-convention */
    'grpc.max_receive_message_length': GRPC_MAX_MESSAGE_LENGTH,
    'grpc.max_send_message_length': GRPC_MAX_MESSAGE_LENGTH,
    /* eslint-enable @typescript-eslint/naming-convention */
  });

  const pfsService = {
    listFile: (params: FileObject) => {
      const listFileRequest = new ListFileRequest();
      const file = fileFromObject(params);

      listFileRequest.setFile(file);

      const stream = client.listFile(listFileRequest, credentialMetadata, {
        deadline: Date.now() + RPC_DEADLINE_MS,
      });

      return streamToObjectArray<FileInfo, FileInfo.AsObject>(stream);
    },
    diffFile: (
      newFileObject: FileObject,
      oldFileObject?: FileObject,
      shallow = false,
    ) => {
      const diffFileRequest = new DiffFileRequest();

      const newFile = fileFromObject(newFileObject);
      diffFileRequest.setNewFile(newFile);
      diffFileRequest.setShallow(shallow);

      if (oldFileObject) {
        const oldFile = fileFromObject(oldFileObject);
        diffFileRequest.setOldFile(oldFile);
      }

      const stream = client.diffFile(diffFileRequest, credentialMetadata, {
        deadline: Date.now() + RPC_DEADLINE_MS,
      });

      return streamToObjectArray<DiffFileResponse, DiffFileResponse.AsObject>(
        stream,
      );
    },
    getFile: (params: FileObject, tar = false) => {
      const getFileRequest = new GetFileRequest();
      const file = fileFromObject(params);

      getFileRequest.setFile(file);

      const stream = tar
        ? client.getFileTAR(getFileRequest, credentialMetadata, {
            deadline: Date.now() + RPC_DEADLINE_MS,
          })
        : client.getFile(getFileRequest, credentialMetadata, {
            deadline: Date.now() + RPC_DEADLINE_MS,
          });

      return new Promise<Buffer>((resolve, reject) => {
        const buffers: Buffer[] = [];

        stream.on('data', (chunk: BytesValue) =>
          buffers.push(Buffer.from(chunk.getValue())),
        );
        stream.on('end', () => {
          if (buffers.length) {
            return resolve(Buffer.concat(buffers));
          } else {
            return reject(new Error('File does not exist.'));
          }
        });
        stream.on('error', (err) => reject(err));
      });
    },
    getFileTAR: (params: FileObject) => {
      return pfsService.getFile(params, true);
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
    listCommit: ({
      number,
      all = true,
      originKind,
      from,
      to,
      repo,
      reverse = false,
    }: ListCommitArgs) => {
      const listCommitRequest = new ListCommitRequest();
      if (repo) {
        listCommitRequest.setRepo(repoFromObject(repo).setType('user'));
      }

      if (from) {
        listCommitRequest.setFrom(commitFromObject(from));
      }

      if (to) {
        listCommitRequest.setTo(commitFromObject(to));
      }

      if (number) {
        listCommitRequest.setNumber(number);
      }

      if (originKind) {
        listCommitRequest.setOriginKind(originKind);
      }

      listCommitRequest.setAll(all);
      listCommitRequest.setReverse(reverse);

      const stream = client.listCommit(listCommitRequest, credentialMetadata, {
        deadline: Date.now() + RPC_DEADLINE_MS,
      });

      return streamToObjectArray<CommitInfo, CommitInfo.AsObject>(stream);
    },
    startCommit: ({
      branch,
      parent,
      description = '',
    }: StartCommitRequestArgs) => {
      return new Promise<Commit.AsObject>((resolve, reject) => {
        const startCommitRequest = new StartCommitRequest();

        startCommitRequest.setBranch(branchFromObject(branch));
        if (parent) {
          startCommitRequest.setParent(commitFromObject(parent));
        }
        startCommitRequest.setDescription(description);

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
    finishCommit: ({
      error,
      force = false,
      commit,
      description = '',
    }: FinishCommitRequestArgs) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const finishCommitRequest = new FinishCommitRequest();

        if (error) {
          finishCommitRequest.setError(error);
        }
        if (commit) {
          finishCommitRequest.setCommit(commitFromObject(commit));
        }
        finishCommitRequest.setForce(force);
        finishCommitRequest.setDescription(description);

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
    inspectCommit: ({wait, commit}: InspectCommitRequestArgs) => {
      return new Promise<CommitInfo.AsObject>((resolve, reject) => {
        const inspectCommitRequest = new InspectCommitRequest();

        if (wait) {
          inspectCommitRequest.setWait(wait);
        }

        if (commit) {
          inspectCommitRequest.setCommit(commitFromObject(commit));
        }
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
    subscribeCommit: ({
      repo,
      branch,
      state,
      all = true,
      originKind,
      from,
    }: SubscribeCommitRequestArgs) => {
      const subscribeCommitRequest = new SubscribeCommitRequest();

      subscribeCommitRequest.setRepo(repoFromObject(repo).setType('user'));

      if (from) {
        subscribeCommitRequest.setFrom(commitFromObject(from));
      }

      if (branch) {
        subscribeCommitRequest.setBranch(branch);
      }

      if (state) {
        subscribeCommitRequest.setState(state);
      }

      if (originKind) {
        subscribeCommitRequest.setOriginKind(originKind);
      }

      subscribeCommitRequest.setAll(all);
      const stream = client.subscribeCommit(
        subscribeCommitRequest,
        credentialMetadata,
        {
          deadline: Date.now() + RPC_DEADLINE_MS,
        },
      );

      return streamToObjectArray<CommitInfo, CommitInfo.AsObject>(stream);
    },
    inspectCommitSet: ({commitSet, wait = true}: InspectCommitSetArgs) => {
      const inspectCommitSetRequest = new InspectCommitSetRequest();

      inspectCommitSetRequest.setCommitSet(commitSetFromObject(commitSet));
      inspectCommitSetRequest.setWait(wait);

      const stream = client.inspectCommitSet(
        inspectCommitSetRequest,
        credentialMetadata,
        {
          deadline: Date.now() + RPC_DEADLINE_MS,
        },
      );

      return streamToObjectArray<CommitInfo, CommitInfo.AsObject>(stream);
    },
    listCommitSet: () => {
      const listCommitSetRequest = new ListCommitSetRequest();
      const stream = client.listCommitSet(
        listCommitSetRequest,
        credentialMetadata,
        {
          deadline: Date.now() + RPC_DEADLINE_MS,
        },
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
    dropCommitSet: (commitSet: CommitSetObject) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const request = new DropCommitSetRequest();
        request.setCommitSet(commitSetFromObject(commitSet));
        client.dropCommitSet(request, credentialMetadata, (error) => {
          if (error) return reject(error);
          return resolve({});
        });
      });
    },
    createBranch: ({
      head,
      branch,
      provenance,
      trigger,
      newCommitSet,
    }: CreateBranchArgs) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const createBranchRequest = new CreateBranchRequest();

        if (head) {
          createBranchRequest.setHead(commitFromObject(head));
        }
        if (branch) {
          createBranchRequest.setBranch(branchFromObject(branch));
        }

        if (provenance) {
          const provenanceArray: Branch[] = provenance.map(
            (eachProvenanceObject) => {
              return branchFromObject(eachProvenanceObject);
            },
          );
          createBranchRequest.setProvenanceList(provenanceArray);
        }

        if (trigger) {
          createBranchRequest.setTrigger(triggerFromObject(trigger));
        }
        createBranchRequest.setNewCommitSet(newCommitSet);

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
    listBranch: ({repo, reverse = false}: ListBranchRequestArgs) => {
      const listBranchRequest = new ListBranchRequest();

      listBranchRequest.setRepo(
        new Repo().setName(repo?.name || '').setType('user'),
      );
      listBranchRequest.setReverse(reverse);
      const stream = client.listBranch(listBranchRequest, credentialMetadata, {
        deadline: Date.now() + RPC_DEADLINE_MS,
      });

      return streamToObjectArray<BranchInfo, BranchInfo.AsObject>(stream);
    },
    deleteBranch: ({branch, force = false}: DeleteBranchRequestArgs) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const deleteBranchRequest = new DeleteBranchRequest();

        if (branch) {
          deleteBranchRequest.setBranch(
            new Branch()
              .setName(branch.name)
              .setRepo(
                new Repo().setName(branch.repo?.name || '').setType('user'),
              ),
          );
        }
        deleteBranchRequest.setForce(force);

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
      const stream = client.listRepo(listRepoRequest, credentialMetadata, {
        deadline: Date.now() + RPC_DEADLINE_MS,
      });

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
    createRepo: ({
      repo,
      description = '',
      update = false,
    }: CreateRepoRequestArgs) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const createRepoRequest = new CreateRepoRequest();

        createRepoRequest.setRepo(repoFromObject(repo));
        createRepoRequest.setDescription(description);
        createRepoRequest.setUpdate(update);

        client.createRepo(createRepoRequest, credentialMetadata, (error) => {
          if (error) {
            return reject(error);
          }
          return resolve({});
        });
      });
    },
    deleteRepo: ({repo, force = false}: DeleteRepoRequestArgs) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const deleteRepoRequest = new DeleteRepoRequest();

        deleteRepoRequest.setRepo(repoFromObject(repo));
        deleteRepoRequest.setForce(force);
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
    addFileSet: ({fileSetId, commit}: AddFileSetRequestArgs) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const request = new AddFileSetRequest()
          .setCommit(commitFromObject(commit))
          .setFileSetId(fileSetId);

        client.addFileSet(request, credentialMetadata, (err) => {
          if (err) reject(err);
          else resolve({});
        });
      });
    },
    renewFileSet: ({fileSetId, duration = 600}: RenewFileSetRequestArgs) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const request = new RenewFileSetRequest()
          .setFileSetId(fileSetId)
          .setTtlSeconds(duration);

        client.renewFileSet(request, credentialMetadata, (err) => {
          if (err) reject(err);
          else resolve({});
        });
      });
    },
    composeFileSet: (fileSets: string[]) => {
      return new Promise<string>((resolve, reject) => {
        const request = new ComposeFileSetRequest().setFileSetIdsList(fileSets);

        client.composeFileSet(request, credentialMetadata, (err, res) => {
          if (err) reject(err);
          else resolve(res.getFileSetId());
        });
      });
    },
    modifyFile: async () => {
      return new ModifyFile({
        pachdAddress,
        channelCredentials,
        credentialMetadata,
        plugins,
      });
    },
    fileSet: async () => {
      return new FileSet({
        pachdAddress,
        channelCredentials,
        credentialMetadata,
        plugins,
      });
    },
  };

  return pfsService;
};

export default pfs;
