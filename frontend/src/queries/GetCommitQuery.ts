import {gql} from '@apollo/client';

import {CommitFragment} from '@dash-frontend/fragments/Commit';

export const GET_COMMIT_QUERY = gql`
  query commit($args: CommitQueryArgs!) {
    commit(args: $args) {
      ...CommitFragment
    }
  }
  ${CommitFragment}
`;
