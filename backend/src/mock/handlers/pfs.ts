import fs from 'fs';
import path from 'path';

import {Status} from '@grpc/grpc-js/build/src/constants';
import {
  Permission,
  PfsIAPIServer,
  RepoAuthInfo,
  RepoInfo,
  Branch,
  Repo,
} from '@pachyderm/node-pachyderm';
import {timestampFromObject} from '@pachyderm/node-pachyderm/dist/builders/protobuf';
import {Empty} from 'google-protobuf/google/protobuf/empty_pb';
import {BytesValue} from 'google-protobuf/google/protobuf/wrappers_pb';

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
  let state = {...repos};
  return {
    getService: (): Pick<
      PfsIAPIServer,
      | 'listRepo'
      | 'inspectRepo'
      | 'listCommit'
      | 'listFile'
      | 'createRepo'
      | 'getFile'
    > => {
      return {
        listRepo: (call) => {
          const [projectId] = call.metadata.get('project-id');
          const [accountId] = call.metadata.get('authn-token');
          const projectRepos = setAuthInfoForRepos(
            projectId ? state[projectId.toString()] : state['1'],
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
            projectId ? state[projectId.toString()] : state['tutorial']
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

          const allCommits = commits[projectId.toString()] || commits['1'];

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
            ? files[projectId.toString()]
            : files['1'];
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
            ? state[projectId.toString()]
            : state['1'];
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
      };
    },
    resetState: () => {
      state = {...repos};
    },
  };
};

export default pfs();
