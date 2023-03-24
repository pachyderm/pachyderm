import {gql} from '@apollo/client';

export const DELETE_PROJECT_PSEUDO_FORCE_MUTATION = gql`
  mutation deleteProjectAndResources($args: DeleteProjectAndResourcesArgs!) {
    deleteProjectAndResources(args: $args)
  }
`;

export default DELETE_PROJECT_PSEUDO_FORCE_MUTATION;
