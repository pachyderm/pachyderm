import {gql} from '@apollo/client';

import {JobOverviewFragment} from '@dash-frontend/fragments/JobOverview';

export const JobSetFragment = gql`
  fragment JobSetFields on JobSet {
    id
    state
    createdAt
    finishedAt
    inProgress
    jobs {
      ...JobOverview
      inputString
      inputBranch
      transformString
      transform {
        cmdList
        image
      }
    }
  }
  ${JobOverviewFragment}
`;
