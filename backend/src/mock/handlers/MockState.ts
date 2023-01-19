import {ServiceError} from '@grpc/grpc-js';
import cloneDeep from 'lodash/cloneDeep';

import admin from '@dash-backend/mock/fixtures/admin';
import commits from '@dash-backend/mock/fixtures/commits';
import datums from '@dash-backend/mock/fixtures/datums';
import diffResponses from '@dash-backend/mock/fixtures/diffResponses';
import enterpriseStates from '@dash-backend/mock/fixtures/enterprise';
import files, {Files} from '@dash-backend/mock/fixtures/files';
import jobs from '@dash-backend/mock/fixtures/jobs';
import pipelines from '@dash-backend/mock/fixtures/pipelines';
import {projectInfo as projects} from '@dash-backend/mock/fixtures/projects';
import repos from '@dash-backend/mock/fixtures/repos';
import {
  RepoInfo,
  CommitInfo,
  DatumInfo,
  PipelineInfo,
  JobInfo,
  LogMessage,
  ModifyFileRequest,
  DiffFileResponse,
  GetStateResponse,
  ClusterInfo,
  ProjectInfo,
} from '@dash-backend/proto';

import jobSets from '../fixtures/jobSets';
import {pipelineAndJobLogs, workspaceLogs} from '../fixtures/logs';

export type StateType = {
  repos: {[projectId: string]: RepoInfo[]};
  commits: {[projectId: string]: CommitInfo[]};
  files: Files;
  fileSets: {[projectId: string]: {[fileSetId: string]: ModifyFileRequest[]}};
  diffResponses: {[projectId: string]: {[path: string]: DiffFileResponse}};
  error: ServiceError | null;
  pipelines: {
    [projectId: string]: PipelineInfo[];
  };
  jobs: {
    [projectId: string]: JobInfo[];
  };
  jobSets: {
    [projectId: string]: {
      [id: string]: JobInfo[];
    };
  };
  pipelineAndJobLogs: {
    [projectId: string]: LogMessage[];
  };
  workspaceLogs: LogMessage[];
  projects: {
    [projectId: string]: ProjectInfo;
  };
  datums: {
    [projectId: string]: {[pipelineId: string]: {[jobId: string]: DatumInfo[]}};
  };
  enterprise: GetStateResponse;
  admin: ClusterInfo;
};

const defaultState: StateType = {
  repos,
  commits,
  files,
  fileSets: {},
  diffResponses,
  error: null,
  pipelines,
  jobs,
  jobSets,
  pipelineAndJobLogs,
  workspaceLogs,
  projects,
  datums,
  enterprise: enterpriseStates.active,
  admin: admin,
};

class MockState {
  state: StateType;

  constructor() {
    this.state = {...cloneDeep(defaultState)};
  }

  resetState() {
    this.state = {...cloneDeep(defaultState)};
  }

  getState() {
    return this.state;
  }
}

export default new MockState();
