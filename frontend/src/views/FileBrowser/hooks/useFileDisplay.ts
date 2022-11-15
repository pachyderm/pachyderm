import {File, FileType} from '@graphqlTypes';
import {format, fromUnixTime} from 'date-fns';
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
  'yaml',
];

const useFileDisplay = (file: File) => {
  const {repoId, commitId, branchId, projectId} = useUrlState();
  const filePath = fileBrowserRoute({
    repoId,
    branchId,
    projectId,
    commitId,
    // remove forward slash from path for route
    filePath: file.path.slice(1),
  });

  const fileName =
    file.path
      .split('/')
      .filter((f) => f)
      .pop() || 'unknown';

  const {copy, supported: copySupported} = useClipboardCopy(
    `${repoId}@${branchId}=${file.commitId}:${file.path}`,
  );

  const dateDisplay = useMemo(
    () =>
      file.committed
        ? format(fromUnixTime(file.committed.seconds), 'MMMM d, yyyy')
        : 'N/A',
    [file.committed],
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

  return {
    copy,
    copySupported,
    fileName,
    dateDisplay,
    filePath,
    fileType,
    fileMajorType,
    previewSupported,
  };
};

export default useFileDisplay;
