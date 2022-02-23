import {gql} from '@apollo/client';

export const FINISH_COMMIT_MUTATION = gql`
  mutation finishCommit($args: FinishCommitArgs!) {
    finishCommit(args: $args)
  }
`;

export default FINISH_COMMIT_MUTATION;
