import {gql} from '@apollo/client';

import {JobOverviewFragment} from '@dash-frontend/fragments/JobOverview';

export const JOBS_QUERY = gql`
  query jobs($args: JobsQueryArgs!) {
    jobs(args: $args) {
      items {
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
      cursor {
        seconds
        nanos
      }
      hasNextPage
    }
  }
  ${JobOverviewFragment}
`;
