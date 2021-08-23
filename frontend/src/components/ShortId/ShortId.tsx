import {
  CopySVG,
  ButtonLink,
  useClipboardCopy,
  SuccessCheckmark,
  Tooltip,
} from '@pachyderm/components';
import classnames from 'classnames';
import React, {useCallback} from 'react';

import styles from './ShortId.module.css';

interface ShortIdProps {
  inputString: string;
}

const ShortId: React.FC<ShortIdProps> = ({inputString}) => {
  const shortId = inputString.slice(0, 8);
  inputString.replace(/[a-zA-Z0-9_-]{64}|[a-zA-Z0-9_-]{32}/, 'ID');
  const {copy, copied, reset} = useClipboardCopy(inputString);

  const handleCopy = useCallback(() => {
    copy();
    setTimeout(reset, 1000);
  }, [copy, reset]);

  return (
    <Tooltip
      tooltipText={`Click to copy: ${inputString}`}
      placement="left"
      tooltipKey={inputString}
    >
      <span className={styles.base}>
        <ButtonLink onClick={handleCopy} data-testid={`CommitPath_copy`}>
          {`${shortId} `}
          <CopySVG
            className={classnames(styles.icon, {[styles.hide]: copied})}
          />
          <SuccessCheckmark
            show={copied}
            className={styles.icon}
            aria-label={'You have successfully copied the id'}
          />
        </ButtonLink>
      </span>
    </Tooltip>
  );
};

export default ShortId;
