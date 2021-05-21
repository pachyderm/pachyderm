import {gql} from '@apollo/client';

export const PIPELINE_JOBS_QUERY = gql`
  query pipelineJobs($args: PipelineJobsQueryArgs!) {
    pipelineJobs(args: $args) {
      id
      state
      createdAt
    }
  }
`;
