import {gql} from '@apollo/client';

import {JobOverviewFragment} from '@dash-frontend/fragments/JobOverview';

export const JOB_SET_QUERY = gql`
  query jobSet($args: JobSetQueryArgs!) {
    jobSet(args: $args) {
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
