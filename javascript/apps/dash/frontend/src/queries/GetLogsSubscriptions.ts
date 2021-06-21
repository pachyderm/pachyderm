import {gql} from '@apollo/client';

import {LogFragment} from '@dash-frontend/fragments/Log';

export const GET_WORKSPACE_LOGS_SUBSCRIPTION = gql`
  subscription getWorkspaceLogStream($args: WorkspaceLogsArgs!) {
    workspaceLogs(args: $args) {
      ...LogFields
    }
  }
  ${LogFragment}
`;

export const GET_PIPELINE_LOGS_SUBSCRIPTION = gql`
  subscription getPipelineLogStream($args: PipelineLogsArgs!) {
    pipelineLogs(args: $args) {
      ...LogFields
    }
  }
  ${LogFragment}
`;

export const GET_JOB_LOGS_SUBSCRIPTION = gql`
  subscription getJobLogStream($args: JobLogsArgs!) {
    jobLogs(args: $args) {
      ...LogFields
    }
  }
  ${LogFragment}
`;
