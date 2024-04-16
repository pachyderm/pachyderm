import React from 'react';

import {
  CaptionTextSmall,
  Icon,
  InfoSVG,
  ArrowUpSVG,
  ArrowDownSVG,
  ArrowRightSVG,
  ArrowLeftSVG,
  Tooltip,
} from '@pachyderm/components';

import styles from './FileShortcuts.module.css';

const FileShortcutsText = () => {
  return (
    <div className={styles.text}>
      <div className={styles.row}>
        Select file{' '}
        <span className={styles.arrow}>
          <ArrowUpSVG /> or <ArrowDownSVG />
        </span>
      </div>
      <div className={styles.row}>
        Open folder <ArrowRightSVG />
      </div>
      <div className={styles.row}>
        Close folder <ArrowLeftSVG /> <ArrowLeftSVG />
      </div>
    </div>
  );
};

const FileShortcuts = () => {
  return (
    <div className={styles.shortcuts}>
      <Tooltip
        tooltipText={<FileShortcutsText />}
        allowedPlacements={['bottom']}
      >
        <CaptionTextSmall>Shortcuts</CaptionTextSmall>
        <Icon small className={styles.info}>
          <InfoSVG />
        </Icon>
      </Tooltip>
    </div>
  );
};

export default FileShortcuts;
