import fs from 'fs';
import path from 'path';

import {ServiceError} from '@grpc/grpc-js';
import {Status} from '@grpc/grpc-js/build/src/constants';
import {Empty} from 'google-protobuf/google/protobuf/empty_pb';
import {BytesValue} from 'google-protobuf/google/protobuf/wrappers_pb';
import uniqueId from 'lodash/uniqueId';

import {REPO_READER_PERMISSIONS} from '@dash-backend/constants/permissions';
import {
  Permission,
  PfsIAPIServer,
  AuthInfo,
  RepoInfo,
  Branch,
  Repo,
  BranchInfo,
  ModifyFileRequest,
  Commit,
  FileType,
  OriginKind,
  CreateFileSetResponse,
  ProjectInfo,
  Project,
  DeleteReposResponse,
} from '@dash-backend/proto';
import {
  commitInfoFromObject,
  fileInfoFromObject,
} from '@dash-backend/proto/builders/pfs';
import {timestampFromObject} from '@dash-backend/proto/builders/protobuf';
import {DeleteRepoResponse} from '@dash-backend/proto/proto/pfs/pfs_pb';
import {createServiceError} from '@dash-backend/testHelpers';

import repoAuthInfos from '../fixtures/authInfos';

import MockState from './MockState';

const setAuthInfoForRepo = (repo: RepoInfo, accountId = '') => {
  const repoName = repo.getRepo()?.getName();
  const authInfo = repoAuthInfos[accountId] || repoAuthInfos.default;

  repo.setAuthInfo(
    repoName && authInfo[repoName]
      ? authInfo[repoName]
      : new AuthInfo().setPermissionsList(REPO_READER_PERMISSIONS),
  );

  return repo;
};

const setAuthInfoForRepos = (repos: RepoInfo[], accountId = '') => {
  repos.forEach((repo) => {
    setAuthInfoForRepo(repo, accountId);
  });

  return repos;
};

const pfs = () => {
  return {
    getService: (): Pick<
      PfsIAPIServer,
      | 'listRepo'
      | 'inspectRepo'
      | 'inspectCommit'
      | 'listCommit'
      | 'findCommits'
      | 'listFile'
      | 'diffFile'
      | 'createRepo'
      | 'getFile'
      | 'createBranch'
      | 'inspectBranch'
      | 'listBranch'
      | 'modifyFile'
      | 'startCommit'
      | 'finishCommit'
      | 'deleteRepo'
      | 'deleteRepos'
      | 'createFileSet'
      | 'addFileSet'
      | 'listProject'
      | 'inspectProject'
      | 'createProject'
      | 'deleteProject'
    > => {
      return {
        listRepo: (call) => {
          // TODO: implement finding all repos across all given projects.
          // If no projects are supplied, search all projects.
          // If projects are supplied, only search those projects.
          const [projectId] = call.metadata.get('project-id');
          const [accountId] = call.metadata.get('authn-token');
          const projectRepos = setAuthInfoForRepos(
            projectId
              ? MockState.state.repos[projectId.toString()] || []
              : MockState.state.repos['Solar-Panel-Data-Sorting'],
            accountId.toString(),
          );

          projectRepos.forEach((repo) => {
            call.write(repo);
          });

          call.end();
        },
        inspectRepo: (call, callback) => {
          const [projectId] = call.metadata.get('project-id');
          const [accountId] = call.metadata.get('authn-token');
          const repoName = call.request.getRepo()?.getName();
          const repo = (
            projectId
              ? MockState.state.repos[projectId.toString()]
              : MockState.state.repos['tutorial']
          ).find((r) => r.getRepo()?.getName() === repoName);

          if (repo) {
            setAuthInfoForRepo(repo, accountId.toString());
            callback(null, repo);
          } else {
            callback({code: Status.NOT_FOUND, details: 'repo not found'});
          }
        },
        inspectCommit: (call, callback) => {
          const [projectId] = call.metadata.get('project-id');
          const [accountId] = call.metadata.get('authn-token');
          const commitId = call.request.getCommit()?.getId();
          const repo = call.request.getCommit()?.getBranch()?.getRepo();

          const projectCommits = projectId
            ? MockState.state.commits[projectId.toString()]
            : MockState.state.commits['Data-Cleaning-Process'];
          const commit = commitId
            ? projectCommits.find((c) => c.getCommit()?.getId() === commitId)
            : projectCommits[0];

          const authInfo =
            repoAuthInfos[accountId.toString()] || repoAuthInfos.default;

          if (repo && accountId) {
            const authRepoInfo = authInfo[repo.getName()];

            if (
              authRepoInfo &&
              !authRepoInfo
                .getPermissionsList()
                .includes(Permission.REPO_LIST_COMMIT)
            ) {
              call.emit(
                'error',
                createServiceError({
                  code: Status.UNKNOWN,
                  details: 'not authorized',
                }),
              );
            }
          }

          if (commit) {
            callback(null, commit);
          } else {
            callback({code: Status.NOT_FOUND, details: 'commit not found'});
          }
        },
        listCommit: (call) => {
          const [projectId] = call.metadata.get('project-id');
          const [accountId] = call.metadata.get('authn-token');

          const repoName = call.request.getRepo()?.getName();
          const branchName = call.request.getTo()?.getBranch()?.getName();
          const originKind = call.request.getOriginKind();

          // thrown by core
          if (
            branchName &&
            call.request.getTo()?.getBranch()?.getRepo()?.getName() !== repoName
          ) {
            call.emit(
              'error',
              createServiceError({
                code: Status.INVALID_ARGUMENT,
                details: `to repo needs to match ${repoName}`,
              }),
            );
          }

          const authInfo =
            repoAuthInfos[accountId.toString()] || repoAuthInfos.default;

          if (repoName && accountId) {
            const authRepoInfo = authInfo[repoName];

            if (
              authRepoInfo &&
              !authRepoInfo
                .getPermissionsList()
                .includes(Permission.REPO_LIST_COMMIT)
            ) {
              call.emit(
                'error',
                createServiceError({
                  code: Status.UNKNOWN,
                  details: 'not authorized',
                }),
              );
            }
          }

          const allCommits =
            MockState.state.commits[projectId.toString()] ||
            MockState.state.commits['Solar-Panel-Data-Sorting'];

          let commits = allCommits.filter((commit) => {
            if (
              commit.getCommit()?.getBranch()?.getRepo()?.getName() ===
                repoName &&
              (!branchName ||
                commit.getCommit()?.getBranch()?.getName() === branchName) &&
              (!originKind || commit.getOrigin()?.getKind() === originKind)
            ) {
              return commit;
            }
          });

          const reverse = call.request.getReverse();
          if (reverse) {
            commits = commits.reverse();
          }

          const cursor = call.request.getStartedTime();
          if (cursor) {
            const cursorIndex = commits.findIndex(
              (commit) =>
                commit.getStarted()?.getNanos() === cursor.getNanos() &&
                commit.getStarted()?.getSeconds() === cursor.getSeconds(),
            );
            commits = commits.slice(cursorIndex !== -1 ? cursorIndex + 1 : -1);
          }

          const number = call.request.getNumber();
          if (number) {
            commits = commits.slice(0, number);
          }
          commits.forEach((commit) => call.write(commit));

          call.end();
        },
        findCommits: (call) => {
          call.end();
        },
        listFile: (call) => {
          const [projectId] = call.metadata.get('project-id');
          const path = call.request.getFile()?.getPath() || '/';
          const directories = projectId
            ? MockState.state.files[projectId.toString()]
            : MockState.state.files['Solar-Panel-Data-Sorting'];
          let replyFiles = directories[path] || directories['/'];

          const reverse = call.request.getReverse();
          if (reverse) {
            replyFiles = replyFiles.reverse();
          }

          const cursor = call.request.getPaginationmarker();
          if (cursor) {
            cursor.getPath;
            const cursorIndex = replyFiles.findIndex(
              (file) => file.getFile()?.getPath() === cursor.getPath(),
            );
            replyFiles = replyFiles.slice(
              cursorIndex !== -1 ? cursorIndex + 1 : -1,
            );
          }

          const number = call.request.getNumber();
          if (number) {
            replyFiles = replyFiles.slice(0, number);
          }

          replyFiles.forEach((file) => call.write(file));
          call.end();
        },
        getFile: (call) => {
          const filePath = call.request.getFile()?.getPath();
          const file = fs.readFileSync(
            path.resolve(__dirname, `../fixtures/files/${filePath}`),
          );
          const bytes = new BytesValue();
          bytes.setValue(file);
          call.write(bytes);
          call.end();
        },
        diffFile: (call) => {
          const [projectId] = call.metadata.get('project-id');
          const path = '/';
          const diff = projectId
            ? MockState.state.diffResponses[projectId.toString()]
            : MockState.state.diffResponses.default;

          call.write(diff[path]);
          call.end();
        },
        createRepo: (call, callback) => {
          const [projectId] = call.metadata.get('project-id');
          const repoName = call.request.getRepo()?.getName();
          const update = call.request.getUpdate();
          const description = call.request.getDescription();
          const projectRepos = projectId
            ? MockState.state.repos[projectId.toString()]
            : MockState.state.repos['Solar-Panel-Data-Sorting'];
          if (repoName) {
            const existingRepo = projectRepos.find(
              (repo) => repo.getRepo()?.getName() === repoName,
            );
            if (existingRepo) {
              if (!update) {
                callback({
                  code: Status.ALREADY_EXISTS,
                  details: `repo ${repoName} already exists`,
                });
              } else {
                existingRepo
                  .setRepo(
                    new Repo()
                      .setName(repoName)
                      .setType(existingRepo.getRepo()?.getType() || 'user'),
                  )
                  .setDescription(description);
              }
            } else {
              const newRepo = new RepoInfo()
                .setRepo(new Repo().setName(repoName).setType('user'))
                .setDetails(new RepoInfo.Details().setSizeBytes(0))
                .setBranchesList([new Branch().setName('master')])
                .setCreated(
                  timestampFromObject({
                    seconds: Math.floor(Date.now() / 1000),
                    nanos: 0,
                  }),
                )
                .setDescription(description);
              projectRepos.push(newRepo);
            }
          }
          callback(null, new Empty());
        },
        deleteRepo: (call, callback) => {
          const [projectId] = call.metadata.get('project-id');
          const request = call.request;
          const projectRepos = MockState.state.repos[projectId.toString()];
          if (!projectRepos) {
            callback({
              code: Status.NOT_FOUND,
              details: `project ${projectId.toString()} does not exist`,
            });
            return;
          }
          const deleteRepo = request.getRepo();
          MockState.state.repos[projectId.toString()] = projectRepos.filter(
            (repo) => {
              return deleteRepo?.getName() !== repo.getRepo()?.getName();
            },
          );

          callback(null, new DeleteRepoResponse().setDeleted(true));
        },
        deleteRepos: (call, callback) => {
          const projects = call.request.getProjectsList();
          if (!projects.some((project) => !!project)) {
            callback({
              code: Status.UNKNOWN,
              details: `no projects were provided`,
            });
            return;
          }
          for (const project of projects) {
            delete MockState.state.repos[project.toString()];
          }

          callback(null, new DeleteReposResponse());
        },
        inspectBranch: (call, callback) => {
          const [projectId] = call.metadata.get('project-id');
          const requestBranch = call.request.getBranch();
          const projectRepos = MockState.state.repos[projectId.toString()];
          if (!projectRepos) {
            callback({
              code: Status.NOT_FOUND,
              details: `project ${projectId.toString()} does not exist`,
            });
            return;
          }
          const repo = projectRepos.find(
            (repo) =>
              repo.getRepo()?.getName() === requestBranch?.getRepo()?.getName(),
          );
          if (!repo) {
            callback({
              code: Status.NOT_FOUND,
              details: `repo ${
                requestBranch?.getRepo()?.getName
              } does not exist`,
            });
            return;
          }
          const branch = repo
            .getBranchesList()
            .find((branch) => branch.getName() === requestBranch?.getName());
          if (!repo) {
            callback({
              code: Status.NOT_FOUND,
              details: `branch ${requestBranch?.getName()} does not exist`,
            });
            return;
          }
          const projectCommits =
            MockState.state.commits[projectId.toString()] || [];
          const head = projectCommits.find(
            (commitInfo) =>
              commitInfo.getCommit()?.getBranch()?.getName() ===
              requestBranch?.getName(),
          );
          const branchInfo = new BranchInfo()
            .setBranch(branch)
            .setHead(head?.getCommit());
          callback(null, branchInfo);
        },
        listBranch: (call) => {
          const projectId =
            call.request.getRepo()?.getProject()?.getName() || 'default';
          const requestBranch = call.request.getRepo();
          const projectRepos = MockState.state.repos[projectId.toString()];
          if (!projectRepos) {
            call.emit(
              'error',
              createServiceError({
                code: Status.NOT_FOUND,
                details: `project ${projectId.toString()} does not exist`,
              }),
            );
            return;
          }
          const repo = projectRepos.find(
            (repo) => repo.getRepo()?.getName() === requestBranch?.getName(),
          );
          if (!repo) {
            call.emit(
              'error',
              createServiceError({
                code: Status.NOT_FOUND,
                details: `repo ${requestBranch?.getName()} does not exist`,
              }),
            );
            return;
          }
          const branches = repo.getBranchesList();

          branches.forEach((branch) => {
            const projectCommits =
              MockState.state.commits[projectId.toString()] || [];
            const head = projectCommits.find(
              (commitInfo) =>
                commitInfo.getCommit()?.getBranch()?.getName() ===
                requestBranch?.getName(),
            );
            const branchInfo = new BranchInfo()
              .setBranch(branch)
              .setHead(head?.getCommit());

            call.write(branchInfo);
          });

          call.end();
        },
        createBranch: (call, callback) => {
          try {
            const [projectId] = call.metadata.get('project-id');
            const requestBranch = call.request.getBranch();
            if (requestBranch) {
              const branchName = requestBranch.getName();
              const repoName = requestBranch.getRepo()?.getName();
              const projectRepos = MockState.state.repos[projectId.toString()];
              if (!projectRepos) {
                callback({
                  code: Status.NOT_FOUND,
                  details: `project ${projectId.toString()} does not exist`,
                });
                return;
              }
              const repo = projectRepos.find(
                (repoInfo) => repoInfo.getRepo()?.getName() === repoName,
              );
              if (!repo) {
                callback({
                  code: Status.NOT_FOUND,
                  details: `repo ${repoName} does not exist`,
                });
              }
              const branches = repo?.getBranchesList();

              if (call.request.getHead()) {
                const commitId = call.request.getHead()?.getId();
                const commitIndex =
                  commitId === '^'
                    ? 0
                    : MockState.state.commits[projectId.toString()].findIndex(
                        (commit) =>
                          commit.getCommit()?.getId() ===
                          call.request.getHead()?.getId(),
                      );
                if (commitIndex === -1) {
                  callback({
                    code: Status.NOT_FOUND,
                    details: `commit ${call.request
                      .getHead()
                      ?.getId()} does not exist`,
                  });
                  return;
                }
                MockState.state.commits[projectId.toString()] =
                  MockState.state.commits[projectId.toString()].filter(
                    (commit, index) => {
                      return (
                        commit.getCommit()?.getBranch()?.getName() ===
                          branchName && index <= commitIndex
                      );
                    },
                  );
              }
              const existingBranch = branches?.find(
                (b) => b.getName() === branchName,
              );
              if (!existingBranch) {
                branches?.push(
                  new Branch()
                    .setName(branchName || '')
                    .setRepo(call.request.getBranch()?.getRepo()),
                );
              }
            }

            callback(null, new Empty());
          } catch (e) {
            console.error(e);
          }
        },
        modifyFile: (call, callback) => {
          const [projectId] = call.metadata.get('project-id');
          let commit: Commit | undefined;
          let autoCommit = false;

          call.on('error', (err) => {
            callback(err as ServiceError);
            autoCommit = false;
          });
          call.on('end', () => {
            const commitInfo = MockState.state.commits[
              projectId.toString()
            ].find((commitInfo) => {
              return commitInfo.getCommit()?.getId() === commit?.getId();
            });
            if (autoCommit && commitInfo) {
              commitInfo.setFinished(
                timestampFromObject({
                  seconds: Math.ceil(Date.now() / 1000),
                  nanos: 0,
                }),
              );
            }
            autoCommit = false;
            callback(null, new Empty());
          });
          call.on('data', async (chunk: ModifyFileRequest) => {
            if (!commit) {
              commit = chunk.hasSetCommit() ? chunk.getSetCommit() : commit;
              if (commit) {
                if (!MockState.state['commits'][projectId.toString()]) {
                  MockState.state['commits'][projectId.toString()] = [];
                }
                if (
                  !MockState.state['commits'][projectId.toString()].find(
                    (commitInfo) =>
                      commitInfo.getCommit()?.getId() === commit?.getId(),
                  ) &&
                  !commit?.getBranch()
                ) {
                  call.emit(
                    'error',
                    createServiceError({
                      code: Status.CANCELLED,
                      details: 'Commit must have a branch',
                    }),
                  );
                } else {
                  commit.setId(uniqueId());
                  autoCommit = true;
                  MockState.state['commits'][projectId.toString()].push(
                    commitInfoFromObject({
                      commit: commit.toObject(),
                      sizeBytes: 0,
                      description: '',
                      started: {
                        seconds: Math.ceil(Date.now() / 1000),
                        nanos: 0,
                      },
                      originKind: OriginKind.USER,
                    }),
                  );
                }
              } else {
                call.emit(
                  'error',
                  createServiceError({
                    code: Status.CANCELLED,
                    details: 'Commit must be set before adding files',
                  }),
                );
              }
            }
            const addFile = chunk.getAddFile();
            const deleteFile = chunk.getDeleteFile();

            if (addFile && commit) {
              const url = addFile.getUrl()?.getUrl();
              const raw = addFile.getRaw();
              if (url || raw) {
                const sizeBytes = raw
                  ? raw.getValue().length
                  : Math.floor(Math.random() * 1200);
                const fileInfo = fileInfoFromObject({
                  committed: {
                    seconds: Math.ceil(Date.now() / 1000),
                    nanos: 0,
                  },
                  file: {
                    commitId: commit.getId(),
                    path: addFile.getPath(),
                    branch: {
                      name: commit.getBranch()?.getName() || 'master',
                      repo: commit.getBranch()?.getRepo()?.toObject(),
                    },
                  },
                  fileType: FileType.FILE,
                  hash: uniqueId(),
                  sizeBytes,
                });
                const dirPath = path.dirname(addFile.getPath());
                if (!MockState.state.files[projectId.toString()]) {
                  MockState.state.files[projectId.toString()] = {};
                }
                if (
                  MockState.state.files[projectId.toString()][dirPath] &&
                  !MockState.state.files[projectId.toString()][dirPath].some(
                    (file) => file.getFile()?.getPath() === addFile.getPath(),
                  )
                ) {
                  MockState.state.files[projectId.toString()][dirPath].push(
                    fileInfo,
                  );
                } else {
                  MockState.state.files[projectId.toString()] = {
                    ...MockState.state.files[projectId.toString()],
                    [dirPath]: [fileInfo],
                  };
                }
                const commitInfo = MockState.state.commits[
                  projectId.toString()
                ].find((commitInfo) => {
                  return commitInfo.getCommit()?.getId() === commit?.getId();
                });
                commitInfo
                  ?.getDetails()
                  ?.setSizeBytes(
                    (commitInfo.getDetails()?.getSizeBytes() || 0) + sizeBytes,
                  );
              } else {
                call.emit(
                  'error',
                  createServiceError({
                    code: Status.INVALID_ARGUMENT,
                    details: 'File data must be specified',
                  }),
                );
              }
            }
            if (deleteFile && commit) {
              const pathToDelete = deleteFile.getPath();
              const dirPath = path.dirname(pathToDelete);
              const projectFiles = MockState.state.files[projectId.toString()];

              if (projectFiles && projectFiles[dirPath]) {
                MockState.state.files[projectId.toString()] = {
                  ...projectFiles,
                  [dirPath]: projectFiles[dirPath].filter((file) => {
                    return pathToDelete !== file.getFile()?.getPath();
                  }),
                };
              }
            }
          });
        },
        createFileSet: (call, callback) => {
          const [projectId] = call.metadata.get('project-id');
          const fileSetId = uniqueId();
          call.on('error', (err) => {
            callback(err as ServiceError);
          });
          call.on('end', () => {
            callback(null, new CreateFileSetResponse().setFileSetId(fileSetId));
          });
          call.on('data', async (chunk: ModifyFileRequest) => {
            if (!MockState.state.fileSets[projectId.toString()]) {
              MockState.state.fileSets[projectId.toString()] = {};
            }
            if (MockState.state.fileSets[projectId.toString()][fileSetId]) {
              MockState.state.fileSets[projectId.toString()][fileSetId].push(
                chunk,
              );
            } else {
              MockState.state.fileSets[projectId.toString()][fileSetId] = [
                chunk,
              ];
            }
          });
        },
        addFileSet: (call, callback) => {
          const [projectId] = call.metadata.get('project-id');
          const fileSetId = call.request.getFileSetId();
          const commit = call.request.getCommit();

          if (!commit) {
            call.emit(
              'error',
              createServiceError({
                code: Status.CANCELLED,
                details: 'Commit must be set',
              }),
            );
          }

          if (MockState.state.fileSets[projectId.toString()][fileSetId]) {
            MockState.state.fileSets[projectId.toString()][fileSetId].forEach(
              (request) => {
                const addFile = request.getAddFile();
                const deleteFile = request.getDeleteFile();

                if (addFile && commit) {
                  const url = addFile.getUrl()?.getUrl();
                  if (url) {
                    const sizeBytes = Math.floor(Math.random() * 1200);
                    const fileInfo = fileInfoFromObject({
                      committed: {
                        seconds: Math.ceil(Date.now() / 1000),
                        nanos: 0,
                      },
                      file: {
                        commitId: commit.getId(),
                        path: addFile.getPath(),
                        branch: {
                          name: commit.getBranch()?.getName() || 'master',
                          repo: commit.getBranch()?.getRepo()?.toObject(),
                        },
                      },
                      fileType: FileType.FILE,
                      hash: uniqueId(),
                      sizeBytes,
                    });
                    const dirPath = path.dirname(addFile.getPath());
                    if (!MockState.state.files[projectId.toString()]) {
                      MockState.state.files[projectId.toString()] = {};
                    }
                    if (
                      MockState.state.files[projectId.toString()][dirPath] &&
                      !MockState.state.files[projectId.toString()][
                        dirPath
                      ].some(
                        (file) =>
                          file.getFile()?.getPath() === addFile.getPath(),
                      )
                    ) {
                      MockState.state.files[projectId.toString()][dirPath].push(
                        fileInfo,
                      );
                    } else {
                      MockState.state.files[projectId.toString()] = {
                        ...MockState.state.files[projectId.toString()],
                        [dirPath]: [fileInfo],
                      };
                    }
                    const commitInfo = MockState.state.commits[
                      projectId.toString()
                    ].find((commitInfo) => {
                      return (
                        commitInfo.getCommit()?.getId() === commit?.getId()
                      );
                    });
                    commitInfo
                      ?.getDetails()
                      ?.setSizeBytes(
                        (commitInfo.getDetails()?.getSizeBytes() || 0) +
                          sizeBytes,
                      );
                  } else {
                    call.emit(
                      'error',
                      createServiceError({
                        code: Status.INVALID_ARGUMENT,
                        details: 'URL must be specified',
                      }),
                    );
                  }
                }
                if (deleteFile && commit) {
                  const pathToDelete = deleteFile.getPath();
                  const dirPath = path.dirname(pathToDelete);
                  const projectFiles =
                    MockState.state.files[projectId.toString()];

                  MockState.state.files[projectId.toString()] = {
                    ...projectFiles,
                    [dirPath]: projectFiles[dirPath].filter((file) => {
                      return pathToDelete !== file.getFile()?.getPath();
                    }),
                  };
                }
              },
            );
            delete MockState.state.fileSets[projectId.toString()][fileSetId];
            callback(null, new Empty());
          } else {
            callback({
              code: Status.NOT_FOUND,
              details: `A fileset with id ${fileSetId} does not exist`,
            });
          }
        },
        startCommit: (call, callback) => {
          const [projectId] = call.metadata.get('project-id');
          const request = call.request;
          const newCommit = commitInfoFromObject({
            commit: {
              id: uniqueId(),
              branch: {
                name: request.getBranch()?.getName() || 'master',
                repo: {name: request.getBranch()?.getRepo()?.getName() || ''},
              },
            },
            sizeBytes: 0,
            description: request.getDescription() || '',
            started: {
              seconds: Math.ceil(Date.now() / 1000),
              nanos: 0,
            },
            originKind: OriginKind.USER,
          });
          if (MockState.state.commits[projectId.toString()])
            MockState.state.commits[projectId.toString()].push(newCommit);
          else MockState.state.commits[projectId.toString()] = [newCommit];

          callback(null, newCommit.getCommit());
        },
        finishCommit: (call, callback) => {
          const [projectId] = call.metadata.get('project-id');
          const request = call.request;
          const commit = MockState.state.commits[projectId.toString()].find(
            (commitInfo) => {
              return (
                commitInfo.getCommit()?.getId() === request.getCommit()?.getId()
              );
            },
          );
          if (commit) {
            commit.setFinished(
              timestampFromObject({
                seconds: Math.ceil(Date.now() / 1000),
                nanos: 0,
              }),
            );
            callback(null, new Empty());
          } else {
            callback({
              code: Status.NOT_FOUND,
              details: `A commit with id ${request
                .getCommit()
                ?.getId()} does not exist`,
            });
          }
        },
        listProject: (call) => {
          Object.entries(MockState.state.projects).forEach(
            ([_projectId, project]) => {
              call.write(project);
            },
          );
          call.end();
        },
        inspectProject: (call, callback) => {
          const projectName = call.request.getProject()?.getName();
          const project = Object.values(MockState.state.projects).find(
            (project) => project.getProject()?.getName() === projectName,
          );

          if (project) {
            callback(null, project);
          } else {
            callback({code: Status.NOT_FOUND, details: 'Project not found'});
          }
        },
        createProject: (call, callback) => {
          const projectName = call.request.getProject()?.getName();
          const projectDescription = call.request.getDescription();
          const update = call.request.getUpdate();

          if (!projectName) {
            callback({
              code: Status.UNKNOWN,
              details: `provide a name`,
            });
            return;
          }

          const projectExists =
            projectName && projectName in MockState.state.projects;

          if (!update) {
            if (projectExists) {
              callback({
                code: Status.ALREADY_EXISTS,
                details: `project ${projectName} already exists.`,
              });
              return;
            }

            MockState.state.projects[projectName] = new ProjectInfo()
              .setProject(new Project().setName(projectName))
              .setDescription(projectDescription);

            callback(null, new Empty());
          } else {
            if (!projectExists) {
              callback({
                code: Status.NOT_FOUND,
                details: `project ${projectName} not found`,
              });
              return;
            }

            const project = MockState.state.projects[projectName];
            project.setDescription(projectDescription);
            callback(null, new Empty());
          }
        },
        deleteProject: (call, callback) => {
          const projectName = call.request.getProject()?.getName();
          if (!projectName) {
            callback({
              code: Status.UNKNOWN,
              details: `provide a name`,
            });
            return;
          }
          delete MockState.state.projects[projectName];
          callback(null, new Empty());
        },
      };
    },
  };
};

export default pfs();
