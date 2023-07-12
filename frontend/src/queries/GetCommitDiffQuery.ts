import {gql} from '@apollo/client';

import {DiffFragment} from '@dash-frontend/fragments/Diff';

export const GET_COMMIT_DIFF_QUERY = gql`
  query commitDiff($args: CommitDiffQueryArgs!) {
    commitDiff(args: $args) {
      ...DiffFragment
    }
  }
  ${DiffFragment}
`;
