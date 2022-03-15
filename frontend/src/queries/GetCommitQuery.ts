import {gql} from '@apollo/client';

import {CommitFragment} from '@dash-frontend/fragments/Commit';
import {DiffFragment} from '@dash-frontend/fragments/Diff';

export const GET_COMMIT_QUERY = gql`
  query commit($args: CommitQueryArgs!) {
    commit(args: $args) {
      ...CommitFragment
      diff {
        ...DiffFragment
      }
    }
  }
  ${CommitFragment}
  ${DiffFragment}
`;
