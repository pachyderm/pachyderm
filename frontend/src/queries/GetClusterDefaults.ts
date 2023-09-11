import {gql} from '@apollo/client';

export const GET_CLUSTER_DEFAULTS = gql`
  query GetClusterDefaults {
    getClusterDefaults {
      clusterDefaultsJson
    }
  }
`;
