import {gql} from '@apollo/client';

import {RepoFragment} from '@dash-frontend/fragments/Repo';

export const GET_REPOS_QUERY = gql`
  query repos($args: ReposQueryArgs!) {
    repos(args: $args) {
      ...RepoFragment
    }
  }
  ${RepoFragment}
`;
