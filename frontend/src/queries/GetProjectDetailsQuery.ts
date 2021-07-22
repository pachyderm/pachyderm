import {gql} from '@apollo/client';

import {JobSetFragment} from '@dash-frontend/fragments/JobSet';

export const GET_PROJECT_DETAILS_QUERY = gql`
  query projectDetails($args: ProjectDetailsQueryArgs!) {
    projectDetails(args: $args) {
      sizeDisplay
      repoCount
      pipelineCount
      jobSets {
        ...JobSetFields
      }
    }
  }
  ${JobSetFragment}
`;
