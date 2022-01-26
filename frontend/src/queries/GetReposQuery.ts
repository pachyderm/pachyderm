import {gql} from '@apollo/client';

export const GET_REPOS_QUERY = gql`
  query repos($args: ReposQueryArgs!) {
    repos(args: $args) {
      branches {
        name
      }
      createdAt
      description
      id
      linkedPipeline {
        id
        name
      }
      name
      sizeDisplay
    }
  }
`;
