import {File, FileType} from '@graphqlTypes';
import {useMemo} from 'react';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import getFileMajorType from '@dash-frontend/lib/getFileMajorType';
import {FileMajorType} from '@dash-frontend/lib/types';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
import {useClipboardCopy} from '@pachyderm/components';

const SUPPORTED_PREVIEW_MAJOR_TYPES: FileMajorType[] = [
  'image',
  'video',
  'audio',
];
const SUPPORTED_PREVIEW_MINOR_TYPES: string[] = [
  'json',
  'jsonl',
  'textpb',
  'csv',
  'tsv',
  'tab',
  'html',
  'xml',
  'htm',
  'txt',
  'yml',
  'md',
  'mkd',
  'mdwn',
  'mdown',
  'markdown',
];
const SUPPORTED_VIEW_SOURCE_TYPES: string[] = [
  'md',
  'mkd',
  'mdwn',
  'mdown',
  'markdown',
];

const useFileDisplay = (file: File) => {
  const {repoId, branchId, projectId} = useUrlState();
  const filePath = fileBrowserRoute({
    repoId,
    branchId,
    projectId,
    commitId: file.commitId,
    // remove forward slash from path for route
    filePath: file.path.slice(1),
  });

  const fileName =
    file.path
      .split('/')
      .filter((f) => f)
      .pop() || 'unknown';

  const {copy} = useClipboardCopy(
    `${repoId}@${branchId}=${file.commitId}:${file.path}`,
  );

  const fileType = useMemo(() => {
    if (file.type === FileType.FILE) {
      return file.path.slice(file.path.lastIndexOf('.') + 1).toLowerCase();
    } else if (file.type === FileType.DIR) {
      return 'folder';
    } else {
      return 'unknown';
    }
  }, [file.type, file.path]);

  const fileMajorType = getFileMajorType(fileType);

  const previewSupported = useMemo(() => {
    return (
      SUPPORTED_PREVIEW_MAJOR_TYPES.includes(fileMajorType) ||
      SUPPORTED_PREVIEW_MINOR_TYPES.includes(fileType)
    );
  }, [fileMajorType, fileType]);

  const viewSourceSupported = useMemo(() => {
    return SUPPORTED_VIEW_SOURCE_TYPES.includes(fileType);
  }, [fileType]);

  return {
    copy,
    fileName,
    filePath,
    fileType,
    fileMajorType,
    previewSupported,
    viewSourceSupported,
  };
};

export default useFileDisplay;
