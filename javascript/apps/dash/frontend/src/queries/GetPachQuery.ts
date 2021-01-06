import gql from 'graphql-tag';

export const GET_PACH_QUERY = gql`
  query pach {
    repos {
      name
      isPipelineOutput
    }
    pipelines {
      id
      name
      inputs {
        id
      }
    }
  }
`;
