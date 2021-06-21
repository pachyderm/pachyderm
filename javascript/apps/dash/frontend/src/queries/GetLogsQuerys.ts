import {gql} from '@apollo/client';

import {LogFragment} from '@dash-frontend/fragments/Log';

export const GET_WORKSPACE_LOGS_QUERY = gql`
  query getWorkspaceLogs($args: WorkspaceLogsArgs!) {
    workspaceLogs(args: $args) {
      ...LogFields
    }
  }
  ${LogFragment}
`;

export const GET_PIPELINE_LOGS_QUERY = gql`
  query getPipelineLogs($args: PipelineLogsArgs!) {
    pipelineLogs(args: $args) {
      ...LogFields
    }
  }
  ${LogFragment}
`;

export const GET_JOB_LOGS_QUERY = gql`
  query getJobLogs($args: JobLogsArgs!) {
    jobLogs(args: $args) {
      ...LogFields
    }
  }
  ${LogFragment}
`;
