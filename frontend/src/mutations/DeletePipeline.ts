import {gql} from '@apollo/client';

export const DELETE_PIPELINE_MUTATION = gql`
  mutation deletePipeline($args: DeletePipelineArgs!) {
    deletePipeline(args: $args)
  }
`;

export default DELETE_PIPELINE_MUTATION;
