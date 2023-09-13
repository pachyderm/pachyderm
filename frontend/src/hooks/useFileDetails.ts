import {File} from '@graphqlTypes';
import {useMemo} from 'react';

import {parseFilePath, getFileDetails} from '@dash-frontend/lib/getFileDetails';

const useFileDetails = (path: File['path']) => {
  const {parsedFilePath, fileDetails} = useMemo(() => {
    return {
      fileDetails: getFileDetails(path),
      parsedFilePath: parseFilePath(path),
    };
  }, [path]);

  return {
    fileDetails,
    parsedFilePath,
  };
};

export default useFileDetails;
