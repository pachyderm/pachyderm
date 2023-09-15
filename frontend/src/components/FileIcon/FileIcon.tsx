import React, {HTMLAttributes} from 'react';

import {SupportedFileIcons} from '@dash-frontend/components/CodePreview/getFileDetails';
import {
  FileAudioSVG,
  FileFolderSVG,
  FileImageSVG,
  FileDocSVG,
  FileUnknownSVG,
  FileVideoSVG,
} from '@pachyderm/components';

import styles from './FileIcon.module.css';

interface FileIconProps extends HTMLAttributes<HTMLDivElement> {
  fileType: SupportedFileIcons;
}

const fileIcons = {
  document: <FileDocSVG />,
  image: <FileImageSVG />,
  video: <FileVideoSVG />,
  audio: <FileAudioSVG />,
  folder: <FileFolderSVG />,
  unknown: <FileUnknownSVG />,
};

const FileIcon: React.FC<FileIconProps> = ({className, fileType}) => {
  return (
    <div className={`${styles.base} ${className}`}>{fileIcons[fileType]}</div>
  );
};

export default FileIcon;
