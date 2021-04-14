import {gql} from '@apollo/client';

export const GET_PIPELINE_QUERY = gql`
  query pipeline($args: PipelineQueryArgs!) {
    pipeline(args: $args) {
      id
      name
    }
  }
`;
