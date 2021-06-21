import {gql} from '@apollo/client';

import {JobOverviewFragment} from '@dash-frontend/fragments/JobOverview';

export const JOB_QUERY = gql`
  query job($args: JobQueryArgs!) {
    job(args: $args) {
      ...JobOverview
      inputString
      inputBranch
      transform {
        cmdList
        image
      }
    }
  }
  ${JobOverviewFragment}
`;
