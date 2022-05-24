import {
  CopySVG,
  useClipboardCopy,
  SuccessCheckmark,
  Icon,
  Button,
} from '@pachyderm/components';
import classnames from 'classnames';
import noop from 'lodash/noop';
import React, {useCallback, useState} from 'react';

import styles from './CommitIdCopy.module.css';

interface CommitIdCopyProps {
  repo?: string;
  branch?: string;
  commit: string;
  small?: boolean;
  longId?: boolean;
  clickable?: boolean;
}

const CommitIdCopy: React.FC<CommitIdCopyProps> = ({
  repo,
  branch,
  commit,
  small,
  longId,
  clickable,
}) => {
  const [showIcon, setShowIcon] = useState(!clickable);
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
    <span
      className={classnames(styles.base, {[styles.clickable]: clickable})}
      data-testid={`CommitIdCopy__id`}
      onMouseOver={clickable ? () => setShowIcon(true) : noop}
      onMouseLeave={clickable ? () => setShowIcon(false) : noop}
      onClick={clickable ? handleCopy : noop}
    >
      {!small ? <h5>{copyString}</h5> : copyString}
      {
        <Button
          buttonType="ghost"
          className={classnames(styles.copy, {
            [styles.copied]: copied,
            [styles.show]: showIcon,
          })}
          onClick={handleCopy}
          data-testid={`CommitIdCopy_copy`}
          IconSVG={CopySVG}
        />
      }
      <Icon
        small={small}
        className={classnames(styles.copyCheckmark, {
          [styles.small]: small,
          [styles.copied]: copied,
        })}
      >
        <SuccessCheckmark
          show={copied && showIcon}
          aria-label={'You have successfully copied the id'}
        />
      </Icon>
    </span>
  );
};

export default CommitIdCopy;
