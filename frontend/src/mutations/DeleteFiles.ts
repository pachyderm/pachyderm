import {gql} from '@apollo/client';

export const DELETE_FILES_MUTATION = gql`
  mutation deleteFiles($args: DeleteFilesArgs!) {
    deleteFiles(args: $args)
  }
`;

export default DELETE_FILES_MUTATION;
