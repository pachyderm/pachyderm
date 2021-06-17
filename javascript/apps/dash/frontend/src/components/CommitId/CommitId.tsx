import {
  CopySVG,
  ButtonLink,
  useClipboardCopy,
  SuccessCheckmark,
} from '@pachyderm/components';
import classnames from 'classnames';
import React, {useCallback} from 'react';

import styles from './CommitId.module.css';

interface CommitIdProps {
  commit: string;
}

const CommitId: React.FC<CommitIdProps> = ({commit}) => {
  const shortCommit = commit.slice(0, 8);
  const {copy, copied, reset} = useClipboardCopy(commit);

  const handleCopy = useCallback(() => {
    copy();
    setTimeout(reset, 2000);
  }, [copy, reset]);

  return (
    <span className={styles.base}>
      {shortCommit}

      <ButtonLink onClick={handleCopy} data-testid={`CommitId_copy`}>
        <CopySVG
          className={classnames(styles.info, {[styles.copied]: copied})}
        />
      </ButtonLink>

      <SuccessCheckmark
        show={copied}
        className={styles.copyCheckmark}
        aria-label={'You have successfully copied the shareable url'}
      />
    </span>
  );
};

export default CommitId;
