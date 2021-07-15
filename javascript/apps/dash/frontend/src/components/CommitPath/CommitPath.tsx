import {
  CopySVG,
  ButtonLink,
  useClipboardCopy,
  SuccessCheckmark,
} from '@pachyderm/components';
import classnames from 'classnames';
import React, {useCallback} from 'react';

import styles from './CommitPath.module.css';

interface CommitPathProps {
  repo: string;
  branch: string;
  commit: string;
}

const CommitPath: React.FC<CommitPathProps> = ({repo, branch, commit}) => {
  const shortCommit = commit.slice(0, 8);
  const {copy, copied, reset} = useClipboardCopy(`${repo}@${branch}=${commit}`);

  const handleCopy = useCallback(() => {
    copy();
    setTimeout(reset, 2000);
  }, [copy, reset]);

  return (
    <span className={styles.base}>
      {repo}@{branch}={shortCommit}
      <ButtonLink onClick={handleCopy} data-testid={`CommitPath_copy`}>
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

export default CommitPath;
