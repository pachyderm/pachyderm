import {gql} from '@apollo/client';

import {JobOverviewFragment} from '@dash-frontend/fragments/JobOverview';

export const JobSetFragment = gql`
  fragment JobSetFields on JobSet {
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
  ${JobOverviewFragment}
`;
