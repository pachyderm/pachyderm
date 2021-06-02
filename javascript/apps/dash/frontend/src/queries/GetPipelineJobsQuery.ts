import {gql} from '@apollo/client';

import {PipelineJobOverviewFragment} from '@dash-frontend/fragments/PipelineJobOverview';

export const PIPELINE_JOBS_QUERY = gql`
  query pipelineJobs($args: PipelineJobsQueryArgs!) {
    pipelineJobs(args: $args) {
      ...PipelineJobOverview
      pipelineName
      inputString
      inputBranch
      transform {
        cmdList
        image
      }
    }
  }
  ${PipelineJobOverviewFragment}
`;
