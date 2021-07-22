import {
  FileAudioSVG,
  FileFolderSVG,
  FileImageSVG,
  FileDocSVG,
  FileUnknownSVG,
  FileVideoSVG,
} from '@pachyderm/components';
import React from 'react';

import {FileMajorType} from '@dash-frontend/lib/types';

import styles from './FileIcon.module.css';

type FileIconProps = {
  fileType: FileMajorType;
};

const fileIcons = {
  document: <FileDocSVG />,
  image: <FileImageSVG />,
  video: <FileVideoSVG />,
  audio: <FileAudioSVG />,
  folder: <FileFolderSVG />,
  unknown: <FileUnknownSVG />,
};

const FileIcon: React.FC<FileIconProps> = ({fileType}) => {
  return <div className={styles.base}>{fileIcons[fileType]}</div>;
};

export default FileIcon;
