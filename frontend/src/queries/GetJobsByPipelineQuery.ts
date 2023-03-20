import {gql} from '@apollo/client';

import {JobOverviewFragment} from '@dash-frontend/fragments/JobOverview';

export const JOBS_BY_PIPELINE_QUERY = gql`
  query jobsByPipeline($args: JobsByPipelineQueryArgs!) {
    jobsByPipeline(args: $args) {
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
