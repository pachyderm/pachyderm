import {gql} from '@apollo/client';

export const PUT_FILES_FROM_URLS_MUTATION = gql`
  mutation putFilesFromURLs($args: PutFilesFromURLsArgs!) {
    putFilesFromURLs(args: $args)
  }
`;

export default PUT_FILES_FROM_URLS_MUTATION;
