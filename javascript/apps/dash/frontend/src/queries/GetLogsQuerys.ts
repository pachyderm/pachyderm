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

export const GET_LOGS_QUERY = gql`
  query getLogs($args: LogsArgs!) {
    logs(args: $args) {
      ...LogFields
    }
  }
  ${LogFragment}
`;
