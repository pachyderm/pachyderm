import {gql} from '@apollo/client';

import {PipelineJobOverviewFragment} from '@dash-frontend/fragments/PipelineJobOverview';

export const GET_PROJECT_DETAILS_QUERY = gql`
  query projectDetails($args: ProjectDetailsQueryArgs!) {
    projectDetails(args: $args) {
      sizeDisplay
      repoCount
      pipelineCount
      jobs {
        ...PipelineJobOverview
      }
    }
  }
  ${PipelineJobOverviewFragment}
`;
