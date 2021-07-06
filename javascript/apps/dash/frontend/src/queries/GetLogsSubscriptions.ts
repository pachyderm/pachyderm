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

export const GET_LOGS_SUBSCRIPTION = gql`
  subscription getLogsStream($args: LogsArgs!) {
    logs(args: $args) {
      ...LogFields
    }
  }
  ${LogFragment}
`;
