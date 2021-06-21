import {gql} from '@apollo/client';

import {JobOverviewFragment} from '@dash-frontend/fragments/JobOverview';

export const JOBSET_QUERY = gql`
  query jobset($args: JobsetQueryArgs!) {
    jobset(args: $args) {
      id
      state
      createdAt
      jobs {
        ...JobOverview
        inputString
        inputBranch
        transform {
          cmdList
          image
        }
      }
    }
  }
  ${JobOverviewFragment}
`;
