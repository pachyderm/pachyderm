import {gql} from '@apollo/client';

import {JobOverviewFragment} from '@dash-frontend/fragments/JobOverview';

export const JOBS_QUERY = gql`
  query jobs($args: JobsQueryArgs!) {
    jobs(args: $args) {
      ...JobOverview
      inputString
      inputBranch
      outputBranch
      outputCommit
      reason
      jsonDetails
      transformString
      transform {
        cmdList
        image
        debug
      }
    }
  }
  ${JobOverviewFragment}
`;
