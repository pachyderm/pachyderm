import {gql} from '@apollo/client';

export const RERUN_PIPELINE_MUTATION = gql`
  mutation rerunPipeline($args: RerunPipelineArgs!) {
    rerunPipeline(args: $args)
  }
`;

export default RERUN_PIPELINE_MUTATION;
