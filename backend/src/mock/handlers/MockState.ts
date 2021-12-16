import {ServiceError} from '@grpc/grpc-js';
import {
  RepoInfo,
  CommitInfo,
  PipelineInfo,
  JobInfo,
  LogMessage,
  Project,
  Projects,
} from '@pachyderm/node-pachyderm';
import cloneDeep from 'lodash/cloneDeep';

import commits from '@dash-backend/mock/fixtures/commits';
import files, {Files} from '@dash-backend/mock/fixtures/files';
import jobs from '@dash-backend/mock/fixtures/jobs';
import pipelines from '@dash-backend/mock/fixtures/pipelines';
import {
  allProjects as projects,
  projectInfo,
} from '@dash-backend/mock/fixtures/projects';
import repos from '@dash-backend/mock/fixtures/repos';

import jobSets from '../fixtures/jobSets';
import {pipelineAndJobLogs, workspaceLogs} from '../fixtures/logs';

export type StateType = {
  repos: {[projectId: string]: RepoInfo[]};
  commits: {[projectId: string]: CommitInfo[]};
  files: Files;
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
    [projectId: string]: Project;
  };
  projectInfo: Projects;
};

const defaultState: StateType = {
  repos,
  commits,
  files,
  error: null,
  pipelines,
  jobs,
  jobSets,
  pipelineAndJobLogs,
  workspaceLogs,
  projects,
  projectInfo,
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
