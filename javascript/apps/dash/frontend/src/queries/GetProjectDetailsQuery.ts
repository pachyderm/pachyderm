import {gql} from '@apollo/client';

import {JobOverviewFragment} from '@dash-frontend/fragments/JobOverview';

export const GET_PROJECT_DETAILS_QUERY = gql`
  query projectDetails($args: ProjectDetailsQueryArgs!) {
    projectDetails(args: $args) {
      sizeDisplay
      repoCount
      pipelineCount
      jobs {
        ...JobOverview
      }
    }
  }
  ${JobOverviewFragment}
`;
