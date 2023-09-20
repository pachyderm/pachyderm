import {Empty} from 'google-protobuf/google/protobuf/empty_pb';
import {BytesValue} from 'google-protobuf/google/protobuf/wrappers_pb';

import {timestampFromObject} from '@dash-backend/proto/builders/protobuf';
import {grpcApiConstructorArgs} from '@dash-backend/proto/utils/createGrpcApiClient';

import {
  commitSetFromObject,
  CommitSetObject,
  fileFromObject,
  branchFromObject,
  FileObject,
  repoFromObject,
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
  ListFileArgs,
  StartCommitRequestArgs,
  SubscribeCommitRequestArgs,
  RenewFileSetRequestArgs,
  AddFileSetRequestArgs,
  CreateProjectRequestArgs,
  FindCommitsArgs,
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
  ListProjectRequest,
  Project,
  ProjectInfo,
  InspectProjectRequest,
  CreateProjectRequest,
  DeleteReposRequest,
  DeleteProjectRequest,
  FindCommitsResponse,
  FindCommitsRequest,
  DeleteRepoResponse,
} from '../../proto/pfs/pfs_pb';
import streamToObjectArray from '../../utils/streamToObjectArray';
import {RPC_DEADLINE_MS} from '../constants/rpc';

import {FileSet} from './clients/FileSet';
import {ModifyFile} from './clients/ModifyFile';
import {GRPC_MAX_MESSAGE_LENGTH} from './lib/constants';

let client: APIClient;

const pfsServiceRpcHandler = ({
  credentialMetadata,
  plugins = [],
}: ServiceArgs) => {
  client =
    client ??
    new APIClient(...grpcApiConstructorArgs(), {
      'grpc.max_receive_message_length': GRPC_MAX_MESSAGE_LENGTH,
      'grpc.max_send_message_length': GRPC_MAX_MESSAGE_LENGTH,
    });

  const pfsService = {
    listFile: ({
      projectId,
      commitId,
      cursorPath,
      path,
      branch,
      reverse,
      number,
    }: ListFileArgs) => {
      const listFileRequest = new ListFileRequest();
      const baseFile = fileFromObject({commitId, path, branch});

      baseFile
        .getCommit()
        ?.getBranch()
        ?.getRepo()
        ?.setProject(new Project().setName(projectId));

      listFileRequest.setFile(baseFile);

      if (reverse) {
        listFileRequest.setReverse(reverse);
      }

      if (number) {
        listFileRequest.setNumber(number);
      }

      if (cursorPath) {
        const cursorFile = fileFromObject({commitId, path: cursorPath, branch});

        cursorFile
          .getCommit()
          ?.getBranch()
          ?.getRepo()
          ?.setProject(new Project().setName(projectId));

        listFileRequest.setPaginationmarker(cursorFile);
      }

      const stream = client.listFile(listFileRequest, credentialMetadata, {
        deadline: Date.now() + RPC_DEADLINE_MS,
      });

      return streamToObjectArray<FileInfo, FileInfo.AsObject>(stream);
    },
    diffFile: ({
      projectId,
      newFileObject,
      oldFileObject,
      shallow = false,
    }: {
      projectId: string;
      newFileObject: FileObject;
      oldFileObject?: FileObject;
      shallow?: boolean;
    }) => {
      const diffFileRequest = new DiffFileRequest();

      const newFile = fileFromObject(newFileObject);
      newFile
        .getCommit()
        ?.getBranch()
        ?.getRepo()
        ?.setProject(new Project().setName(projectId));
      diffFileRequest.setNewFile(newFile);
      diffFileRequest.setShallow(shallow);

      if (oldFileObject) {
        const oldFile = fileFromObject(oldFileObject);
        oldFile
          .getCommit()
          ?.getBranch()
          ?.getRepo()
          ?.setProject(new Project().setName(projectId));
        diffFileRequest.setOldFile(oldFile);
      }

      const stream = client.diffFile(diffFileRequest, credentialMetadata, {
        deadline: Date.now() + RPC_DEADLINE_MS,
      });

      return streamToObjectArray<DiffFileResponse, DiffFileResponse.AsObject>(
        stream,
      );
    },
    getFile: ({
      projectId,
      tar = false,
      ...params
    }: {projectId: string; tar?: boolean} & FileObject) => {
      const getFileRequest = new GetFileRequest();
      const file = fileFromObject(params);
      file
        .getCommit()
        ?.getBranch()
        ?.getRepo()
        ?.setProject(new Project().setName(projectId));

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
    getFileTAR: ({
      projectId,
      ...params
    }: {
      projectId: string;
    } & FileObject) => {
      return pfsService.getFile({projectId, ...params, tar: true});
    },
    inspectFile: ({projectId, ...params}: {projectId: string} & FileObject) => {
      return new Promise<FileInfo.AsObject>((resolve, reject) => {
        const inspectFileRequest = new InspectFileRequest();
        const file = fileFromObject(params);
        file
          .getCommit()
          ?.getBranch()
          ?.getRepo()
          ?.setProject(new Project().setName(projectId));

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
      projectId,
      number,
      all = false,
      originKind,
      from,
      to,
      repo,
      reverse = false,
      started_time,
    }: ListCommitArgs) => {
      const listCommitRequest = new ListCommitRequest();
      if (repo) {
        listCommitRequest.setRepo(
          repoFromObject({projectId, ...repo}).setType('user'),
        );
      }

      if (from) {
        listCommitRequest
          .setFrom(commitFromObject(from))
          .getFrom()
          ?.getBranch()
          ?.getRepo()
          ?.setProject(new Project().setName(projectId));
      }

      if (to) {
        listCommitRequest
          .setTo(commitFromObject(to))
          .getTo()
          ?.getBranch()
          ?.getRepo()
          ?.setProject(new Project().setName(projectId));
      }

      if (number) {
        listCommitRequest.setNumber(number);
      }

      if (originKind) {
        listCommitRequest.setOriginKind(originKind);
      }

      listCommitRequest.setAll(all);
      listCommitRequest.setReverse(reverse);

      if (started_time) {
        listCommitRequest.setStartedTime(timestampFromObject(started_time));
      }

      const stream = client.listCommit(listCommitRequest, credentialMetadata, {
        deadline: Date.now() + RPC_DEADLINE_MS,
      });

      return streamToObjectArray<CommitInfo, CommitInfo.AsObject>(stream);
    },
    startCommit: ({
      projectId,
      branch,
      parent,
      description = '',
    }: StartCommitRequestArgs & {projectId: string}) => {
      return new Promise<Commit.AsObject>((resolve, reject) => {
        const startCommitRequest = new StartCommitRequest();

        startCommitRequest.setBranch(branchFromObject(branch));
        startCommitRequest
          .getBranch()
          ?.getRepo()
          ?.setProject(new Project().setName(projectId));

        if (parent) {
          startCommitRequest.setParent(commitFromObject(parent));
          startCommitRequest
            .getParent()
            ?.getBranch()
            ?.getRepo()
            ?.setProject(new Project().setName(projectId));
        }
        startCommitRequest.setDescription(description);
        startCommitRequest.getBranch();

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
      projectId,
      error,
      force = false,
      commit,
      description = '',
    }: FinishCommitRequestArgs & {projectId: string}) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const finishCommitRequest = new FinishCommitRequest();

        if (error) {
          finishCommitRequest.setError(error);
        }
        if (commit) {
          finishCommitRequest.setCommit(commitFromObject(commit));
          finishCommitRequest
            .getCommit()
            ?.getBranch()
            ?.getRepo()
            ?.setProject(new Project().setName(projectId));
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
    inspectCommit: ({
      projectId,
      wait,
      commit,
    }: {projectId: string} & InspectCommitRequestArgs) => {
      return new Promise<CommitInfo.AsObject>((resolve, reject) => {
        const inspectCommitRequest = new InspectCommitRequest();

        if (wait) {
          inspectCommitRequest.setWait(wait);
        }

        if (commit) {
          inspectCommitRequest.setCommit(commitFromObject(commit));
          inspectCommitRequest
            .getCommit()
            ?.getBranch()
            ?.getRepo()
            ?.setProject(new Project().setName(projectId));
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
      projectId,
      repo,
      branch,
      state,
      all = true,
      originKind,
      from,
    }: SubscribeCommitRequestArgs) => {
      const subscribeCommitRequest = new SubscribeCommitRequest();

      subscribeCommitRequest.setRepo(
        repoFromObject({projectId, ...repo}).setType('user'),
      );

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
    findCommits: ({commit, path, limit}: FindCommitsArgs) => {
      const findCommitsRequest = new FindCommitsRequest();
      findCommitsRequest.setStart(commitFromObject(commit));
      findCommitsRequest.setFilePath(path);
      if (limit) {
        findCommitsRequest.setLimit(limit);
      }

      const stream = client.findCommits(
        findCommitsRequest,
        credentialMetadata,
        {
          deadline: Date.now() + RPC_DEADLINE_MS,
        },
      );

      return streamToObjectArray<
        FindCommitsResponse,
        FindCommitsResponse.AsObject
      >(stream);
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
    inspectBranch: ({
      projectId,
      ...params
    }: {projectId: string} & BranchObject) => {
      return new Promise<BranchInfo.AsObject>((resolve, reject) => {
        const inspectBranchRequest = new InspectBranchRequest();
        const branch = branchFromObject(params);
        branch.getRepo()?.setProject(new Project().setName(projectId));

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
    listBranch: ({
      repoName,
      projectId,
      reverse = false,
    }: ListBranchRequestArgs) => {
      const listBranchRequest = new ListBranchRequest();
      const repo = repoFromObject({name: repoName, projectId});
      listBranchRequest.setRepo(repo);
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
    listRepo: ({
      projectIds,
      type = 'user',
    }: {
      projectIds: string[];
      type?: string;
    }) => {
      const listRepoRequest = new ListRepoRequest()
        .setType(type)
        .setProjectsList(
          projectIds.map((projectId) => new Project().setName(projectId)),
        );
      const stream = client.listRepo(listRepoRequest, credentialMetadata, {
        deadline: Date.now() + RPC_DEADLINE_MS,
      });

      return streamToObjectArray<RepoInfo, RepoInfo.AsObject>(stream);
    },
    inspectRepo: ({
      name,
      projectId,
    }: {
      name: Repo.AsObject['name'];
      projectId: string;
    }) => {
      return new Promise<RepoInfo.AsObject>((resolve, reject) => {
        const inspectRepoRequest = new InspectRepoRequest();
        const repo = repoFromObject({name, projectId});
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
      projectId,
      repo,
      description = '',
      update = false,
    }: CreateRepoRequestArgs) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const createRepoRequest = new CreateRepoRequest();

        createRepoRequest.setRepo(repoFromObject({projectId, ...repo}));
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
    deleteRepo: ({projectId, repo, force = false}: DeleteRepoRequestArgs) => {
      return new Promise<DeleteRepoResponse.AsObject>((resolve, reject) => {
        const deleteRepoRequest = new DeleteRepoRequest();

        deleteRepoRequest.setRepo(repoFromObject({projectId, ...repo}));
        deleteRepoRequest.setForce(force);
        client.deleteRepo(
          deleteRepoRequest,
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
    deleteRepos: ({
      projectIds,
      force = false,
    }: {
      projectIds: string[];
      force?: boolean;
    }) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const deleteReposRequest = new DeleteReposRequest();

        const projects = projectIds.map((id) => new Project().setName(id));
        deleteReposRequest.setProjectsList(projects);

        deleteReposRequest.setForce(force);
        client.deleteRepos(deleteReposRequest, credentialMetadata, (error) => {
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
    createProject: ({
      name,
      description,
      update = false,
    }: CreateProjectRequestArgs) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const createProjectRequest = new CreateProjectRequest();

        createProjectRequest.setProject(new Project().setName(name));
        if (description) createProjectRequest.setDescription(description);
        createProjectRequest.setUpdate(update);
        client.createProject(
          createProjectRequest,
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
    deleteProject: ({projectId}: {projectId: string}) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const deleteProjectRequest = new DeleteProjectRequest();
        deleteProjectRequest.setProject(new Project().setName(projectId));
        client.deleteProject(
          deleteProjectRequest,
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
    addFileSet: ({
      projectId,
      fileSetId,
      commit,
    }: AddFileSetRequestArgs & {projectId: string}) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const request = new AddFileSetRequest()
          .setCommit(commitFromObject(commit))
          .setFileSetId(fileSetId);

        request
          .getCommit()
          ?.getBranch()
          ?.getRepo()
          ?.setProject(new Project().setName(projectId));

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
        credentialMetadata,
        plugins,
      });
    },
    fileSet: async () => {
      return new FileSet({
        credentialMetadata,
        plugins,
      });
    },
    listProject: () => {
      const stream = client.listProject(
        new ListProjectRequest(),
        credentialMetadata,
        {
          deadline: Date.now() + RPC_DEADLINE_MS,
        },
      );

      return streamToObjectArray<ProjectInfo, ProjectInfo.AsObject>(stream);
    },
    inspectProject: (id: string) => {
      return new Promise<ProjectInfo.AsObject>((resolve, reject) => {
        const inspectProjectRequest = new InspectProjectRequest();
        inspectProjectRequest.setProject(new Project().setName(id));

        client.inspectProject(
          inspectProjectRequest,
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

  return pfsService;
};

export default pfsServiceRpcHandler;
