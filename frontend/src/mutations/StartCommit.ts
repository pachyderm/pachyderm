import {gql} from '@apollo/client';

export const START_COMMIT_MUTATION = gql`
  mutation startCommit($args: StartCommitArgs!) {
    startCommit(args: $args) {
      branch {
        name
        repo {
          name
          type
        }
      }
      id
    }
  }
`;

export default START_COMMIT_MUTATION;
