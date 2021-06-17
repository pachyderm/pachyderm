import {useClipboardCopy} from '@pachyderm/components';
import {format, fromUnixTime} from 'date-fns';
import {useMemo} from 'react';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {FileMajorType} from '@dash-frontend/lib/types';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
import {File, FileType} from '@graphqlTypes';

const SUPPORTED_PREVIEW_MAJOR_TYPES: FileMajorType[] = [
  'image',
  'video',
  'audio',
];
const SUPPORTED_PREVIEW_MINOR_TYPES: string[] = ['json', 'csv'];

const useFileDisplay = (file: File) => {
  const {repoId, commitId, branchId, projectId} = useUrlState();
  const filePath = fileBrowserRoute({
    repoId,
    branchId,
    projectId,
    commitId,
    filePath: file.path,
  });
  const {copy, supported: copySupported} = useClipboardCopy(filePath);

  const [, fileName] = file.path.split('/');

  const dateDisplay = useMemo(
    () =>
      file.committed
        ? format(fromUnixTime(file.committed.seconds), 'MMMM d, yyyy')
        : 'N/A',
    [file.committed],
  );

  const {fileType, fileMajorType} = useMemo(() => {
    let fileType: string;

    if (file.type === FileType.FILE) {
      fileType = file.path.slice(file.path.lastIndexOf('.') + 1);
    } else if (file.type === FileType.DIR) {
      fileType = 'folder';
    } else {
      fileType = 'unknown';
    }

    let fileMajorType: FileMajorType;

    switch (fileType) {
      case 'pdf':
      case 'xls':
      case 'xlsx':
      case 'html':
      case 'doc':
      case 'docx':
      case 'md':
      case 'csv':
      case 'json':
        fileMajorType = 'document';
        break;
      case 'apng':
      case 'avif':
      case 'gif':
      case 'jpeg':
      case 'png':
      case 'svg':
      case 'webp':
      case 'bmp':
      case 'ico':
      case 'tiff':
        fileMajorType = 'image';
        break;
      case 'mpg':
      case 'mpeg':
      case 'avi':
      case 'wmv':
      case 'mov':
      case 'rm':
      case 'ram':
      case 'swf':
      case 'flv':
      case 'pff':
      case 'webm':
      case 'mp4':
        fileMajorType = 'video';
        break;
      case 'mp3':
      case 'wav':
      case 'ogg':
        fileMajorType = 'audio';
        break;
      case 'folder':
        fileMajorType = 'folder';
        break;
      default:
        fileMajorType = 'unknown';
        break;
    }

    return {fileType, fileMajorType};
  }, [file.path, file.type]);

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
