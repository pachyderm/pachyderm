import {useMemo} from 'react';

import {File} from '@dash-frontend/api/pfs';

import {parseFilePath, getFileDetails} from '../utils/getFileDetails';

const useFileDetails = (path: File['path']) => {
  const {parsedFilePath, fileDetails} = useMemo(() => {
    return {
      fileDetails: getFileDetails(path || ''),
      parsedFilePath: parseFilePath(path || ''),
    };
  }, [path]);

  return {
    fileDetails,
    parsedFilePath,
  };
};

export default useFileDetails;
