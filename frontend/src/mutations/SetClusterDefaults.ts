import {gql} from '@apollo/client';

export const SET_CLUSTER_DEFAULTS = gql`
  mutation setClusterDefaults($args: SetClusterDefaultsArgs!) {
    setClusterDefaults(args: $args) {
      affectedPipelinesList {
        name
        project {
          name
        }
      }
    }
  }
`;

export default SET_CLUSTER_DEFAULTS;
