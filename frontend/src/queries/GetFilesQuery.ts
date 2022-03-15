import {gql} from '@apollo/client';

import {DiffFragment} from '@dash-frontend/fragments/Diff';

export const GET_FILES_QUERY = gql`
  query getFiles($args: FileQueryArgs!) {
    files(args: $args) {
      diff {
        ...DiffFragment
      }
      files {
        committed {
          nanos
          seconds
        }
        commitId
        download
        hash
        path
        repoName
        sizeBytes
        type
        sizeDisplay
        downloadDisabled
        commitAction
      }
    }
  }
  ${DiffFragment}
`;
