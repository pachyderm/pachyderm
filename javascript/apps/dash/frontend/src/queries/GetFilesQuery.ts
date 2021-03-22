import {gql} from '@apollo/client';

export const GET_FILES_QUERY = gql`
  query getFiles($args: FileQueryArgs!) {
    files(args: $args) {
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
    }
  }
`;
