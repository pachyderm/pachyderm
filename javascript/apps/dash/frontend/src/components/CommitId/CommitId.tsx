import {Tooltip, InfoSVG} from '@pachyderm/components';
import React from 'react';

import styles from './CommitId.module.css';

interface CommitIdProps {
  commit: string;
}

const CommitId: React.FC<CommitIdProps> = ({commit}) => {
  const shortCommit = commit.slice(0, 8);
  return (
    <span className={styles.base}>
      {shortCommit}
      <Tooltip
        tooltipText={commit}
        tooltipKey={commit}
        size="large"
        placement="bottom"
      >
        <InfoSVG className={styles.info} />
      </Tooltip>
    </span>
  );
};

export default CommitId;
