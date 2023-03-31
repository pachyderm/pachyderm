import {gql} from '@apollo/client';

import {BranchFragment} from '@dash-frontend/fragments/Branch';

export const GET_BRANCHES_QUERY = gql`
  query getBranches($args: BranchesQueryArgs!) {
    branches(args: $args) {
      ...BranchFragment
    }
  }
  ${BranchFragment}
`;
