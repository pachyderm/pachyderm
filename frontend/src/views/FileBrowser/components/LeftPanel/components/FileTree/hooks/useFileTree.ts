import {useMemo} from 'react';

import {FileType} from '@dash-frontend/api/pfs';
import {useFiles} from '@dash-frontend/hooks/useFiles';
import useUrlState from '@dash-frontend/hooks/useUrlState';

import {MAX_FILES} from '../constants';
import {FileTreeProps} from '../types';

const useFileTree = ({selectedCommitId, path, initial}: FileTreeProps) => {
  const {projectId, repoId, filePath} = useUrlState();
  const filesResult = useFiles(
    {
      projectName: projectId,
      repoName: repoId,
      commitId: selectedCommitId,
      path,
      args: {
        number: MAX_FILES,
      },
    },
    true,
    0,
  );
  const {loading} = filesResult;
  let {files, cursor} = filesResult;

  // Add a root directory at the top level
  if (initial && path === '/') {
    cursor = undefined;
    files = [
      {
        file: {
          path: '/',
        },
        fileType: FileType.DIR,
      },
    ];
  }

  const sortedFiles = useMemo(() => {
    // Show directories first
    return files?.sort((a, b) => {
      if (
        (a.fileType === FileType.DIR && b.fileType === FileType.DIR) ||
        (a.fileType !== FileType.DIR && b.fileType !== FileType.DIR)
      ) {
        // If both are directories or both are files, maintain original order
        return 0;
      } else if (a.fileType === FileType.DIR && b.fileType !== FileType.DIR) {
        // If 'a' is a directory and 'b' is a file, 'a' should come first
        return -1;
      } else {
        // Otherwise, 'b' is a directory and 'a' is a file, so 'b' should come first
        return 1;
      }
    });
  }, [files]);

  return {files: sortedFiles, filePath, projectId, repoId, cursor, loading};
};

export default useFileTree;
