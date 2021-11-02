import fs from 'fs';
import path from 'path';

import {ServiceError} from '@grpc/grpc-js';
import {Status} from '@grpc/grpc-js/build/src/constants';
import {
  Permission,
  PfsIAPIServer,
  RepoAuthInfo,
  RepoInfo,
  Branch,
  Repo,
  ModifyFileRequest,
  Commit,
  FileType,
  OriginKind,
} from '@pachyderm/node-pachyderm';
import {
  commitInfoFromObject,
  fileInfoFromObject,
} from '@pachyderm/node-pachyderm/dist/builders/pfs';
import {timestampFromObject} from '@pachyderm/node-pachyderm/dist/builders/protobuf';
import {Empty} from 'google-protobuf/google/protobuf/empty_pb';
import {BytesValue} from 'google-protobuf/google/protobuf/wrappers_pb';
import uniqueId from 'lodash/uniqueId';

import {REPO_READER_PERMISSIONS} from '@dash-backend/constants/permissions';
import commits from '@dash-backend/mock/fixtures/commits';
import files from '@dash-backend/mock/fixtures/files';
import repos from '@dash-backend/mock/fixtures/repos';
import {createServiceError} from '@dash-backend/testHelpers';

import repoAuthInfos from '../fixtures/repoAuthInfos';

const setAuthInfoForRepo = (repo: RepoInfo, accountId = '') => {
  const repoName = repo.getRepo()?.getName();
  const authInfo = repoAuthInfos[accountId] || repoAuthInfos['default'];

  repo.setAuthInfo(
    repoName && authInfo[repoName]
      ? authInfo[repoName]
      : new RepoAuthInfo().setPermissionsList(REPO_READER_PERMISSIONS),
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
  let state = {repos, commits, files};
  return {
    getService: (): Pick<
      PfsIAPIServer,
      | 'listRepo'
      | 'inspectRepo'
      | 'listCommit'
      | 'listFile'
      | 'createRepo'
      | 'getFile'
      | 'modifyFile'
      | 'startCommit'
      | 'finishCommit'
    > => {
      return {
        listRepo: (call) => {
          const [projectId] = call.metadata.get('project-id');
          const [accountId] = call.metadata.get('authn-token');
          const projectRepos = setAuthInfoForRepos(
            projectId ? state.repos[projectId.toString()] : state.repos['1'],
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
              ? state.repos[projectId.toString()]
              : state.repos['tutorial']
          ).find((r) => r.getRepo()?.getName() === repoName);

          if (repo) {
            setAuthInfoForRepo(repo, accountId.toString());
            callback(null, repo);
          } else {
            callback({code: Status.NOT_FOUND, details: 'repo not found'});
          }
        },
        listCommit: (call) => {
          const [projectId] = call.metadata.get('project-id');
          const [accountId] = call.metadata.get('authn-token');

          const repoName = call.request.getRepo()?.getName();
          const authInfo =
            repoAuthInfos[accountId.toString()] || repoAuthInfos['default'];

          if (repoName && accountId) {
            const authRepoInfo = authInfo[repoName];

            if (
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
            state.commits[projectId.toString()] || state.commits['1'];

          allCommits.forEach((commit) => {
            if (
              commit.getCommit()?.getBranch()?.getRepo()?.getName() === repoName
            ) {
              call.write(commit);
            }
          });

          call.end();
        },
        listFile: (call) => {
          const [projectId] = call.metadata.get('project-id');
          const path = call.request.getFile()?.getPath() || '/';
          const directories = projectId
            ? state.files[projectId.toString()]
            : state.files['1'];
          const replyFiles = directories[path] || directories['/'];
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
        createRepo: (call, callback) => {
          const [projectId] = call.metadata.get('project-id');
          const repoName = call.request.getRepo()?.getName();
          const update = call.request.getUpdate();
          const description = call.request.getDescription();
          const projectRepos = projectId
            ? state.repos[projectId.toString()]
            : state.repos['1'];
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
        modifyFile: (call, callback) => {
          const [projectId] = call.metadata.get('project-id');
          let commit: Commit | undefined;

          call.on('error', (err) => {
            callback(err as ServiceError);
          });
          call.on('end', () => {
            callback(null, new Empty());
          });
          call.on('data', async (chunk: ModifyFileRequest) => {
            commit = chunk.hasSetCommit() ? chunk.getSetCommit() : commit;
            const addFile = chunk.getAddFile();
            if (!commit) {
              call.emit(
                'error',
                createServiceError({
                  code: Status.CANCELLED,
                  details: 'Commit must be set before adding files',
                }),
              );
            }
            if (addFile && commit) {
              const url = addFile.getUrl()?.getUrl();
              if (url) {
                const sizeBytes = Math.floor(Math.random() * 1200);
                const fileInfo = fileInfoFromObject({
                  committed: {seconds: Math.ceil(Date.now() / 1000), nanos: 0},
                  file: {
                    commitId: commit.getId(),
                    path: addFile.getPath(),
                    branch: {
                      name: commit.getBranch()?.getName() || 'master',
                      repo: commit.getBranch()?.getRepo,
                    },
                  },
                  fileType: FileType.FILE,
                  hash: uniqueId(),
                  sizeBytes,
                });
                const dirPath = path.dirname(addFile.getPath());
                if (
                  Object.prototype.hasOwnProperty.call(
                    state.files[projectId.toString()],
                    dirPath,
                  ) &&
                  !state.files[projectId.toString()][dirPath].some(
                    (file) => file.getFile()?.getPath() === addFile.getPath(),
                  )
                ) {
                  state.files[projectId.toString()][dirPath].push(fileInfo);
                } else {
                  state.files[projectId.toString()] = {
                    ...state.files[projectId.toString()],
                    [dirPath]: [fileInfo],
                  };
                }
                const commitInfo = state.commits[projectId.toString()].find(
                  (commitInfo) => {
                    return commitInfo.getCommit()?.getId() === commit?.getId();
                  },
                );
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
                    details: 'URL must be specified',
                  }),
                );
              }
            }
          });
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
          if (state.commits[projectId.toString()])
            state.commits[projectId.toString()].push(newCommit);
          else state.commits[projectId.toString()] = [newCommit];

          callback(null, newCommit.getCommit());
        },
        finishCommit: (call, callback) => {
          const [projectId] = call.metadata.get('project-id');
          const request = call.request;
          const commit = state.commits[projectId.toString()].find(
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
      };
    },
    resetState: () => {
      state = {repos, commits, files};
    },
  };
};

export default pfs();
