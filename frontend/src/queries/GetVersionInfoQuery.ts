import {gql} from '@apollo/client';

export const GET_VERSION_INFO_QUERY = gql`
  query getVersionInfo {
    versionInfo {
      pachdVersion {
        major
        minor
        micro
        additional
        gitCommit
        gitTreeModified
        buildDate
        goVersion
        platform
      }
      consoleVersion
    }
  }
`;
