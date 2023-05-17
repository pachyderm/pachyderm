import {gql} from '@apollo/client';

export const GET_FILE_DOWNLOAD_QUERY = gql`
  query fileDownload($args: FileDownloadArgs!) {
    fileDownload(args: $args)
  }
`;
