import {gql} from '@apollo/client';

import {CommitFragment} from '@dash-frontend/fragments/Commit';

export const COMMIT_SEARCH_QUERY = gql`
  query commitSearch($args: CommitSearchQueryArgs!) {
    commitSearch(args: $args) {
      ...CommitFragment
    }
  }
  ${CommitFragment}
`;
