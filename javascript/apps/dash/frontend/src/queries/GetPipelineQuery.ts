import {gql} from '@apollo/client';

export const GET_PIPELINE_QUERY = gql`
  query pipeline($args: PipelineQueryArgs!) {
    pipeline(args: $args) {
      id
      name
      state
      type
      description
      transform {
        cmdList
        image
      }
      inputString
      cacheSize
      datumTimeoutS
      datumTries
      jobTimeoutS
      enableStats
      outputBranch
      s3OutputRepo
      egress
      schedulingSpec {
        nodeSelectorMap {
          key
          value
        }
        priorityClassName
      }
    }
  }
`;
