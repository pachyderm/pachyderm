import {
  CopySVG,
  ButtonLink,
  useClipboardCopy,
  SuccessCheckmark,
  Icon,
} from '@pachyderm/components';
import classnames from 'classnames';
import React, {useCallback} from 'react';

import styles from './CommitIdCopy.module.css';

interface CommitIdCopyProps {
  repo?: string;
  branch?: string;
  commit: string;
  smallIcon?: boolean;
  longId?: boolean;
}

const CommitIdCopy: React.FC<CommitIdCopyProps> = ({
  repo,
  branch,
  commit,
  smallIcon,
  longId,
}) => {
  const shortCommit = longId ? commit : commit.slice(0, 8);
  const copyString = `${
    repo && branch ? `${repo}@${branch}=` : ''
  }${shortCommit}`;
  const {copy, copied, reset} = useClipboardCopy(copyString);

  const handleCopy = useCallback(() => {
    copy();
    setTimeout(reset, 2000);
  }, [copy, reset]);

  return (
    <span className={styles.base}>
      {copyString}
      <ButtonLink
        className={classnames(styles.copy, {[styles.copied]: copied})}
        onClick={handleCopy}
        data-testid={`CommitIdCopy_copy`}
      >
        <Icon color="plum" small={smallIcon}>
          <CopySVG />
        </Icon>
      </ButtonLink>
      <Icon
        small={smallIcon}
        className={classnames(styles.copyCheckmark, {
          [styles.small]: smallIcon,
          [styles.copied]: copied,
        })}
      >
        <SuccessCheckmark
          show={copied}
          aria-label={'You have successfully copied the id'}
        />
      </Icon>
    </span>
  );
};

export default CommitIdCopy;
