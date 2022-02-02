import {gql} from '@apollo/client';

export const DELETE_FILE_MUTATION = gql`
  mutation deleteFile($args: DeleteFileArgs!) {
    deleteFile(args: $args)
  }
`;

export default DELETE_FILE_MUTATION;
