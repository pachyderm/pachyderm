import {
  CopySVG,
  ButtonLink,
  useClipboardCopy,
  SuccessCheckmark,
  Tooltip,
  Icon,
} from '@pachyderm/components';
import React, {useCallback} from 'react';

import styles from './ShortId.module.css';

interface ShortIdProps {
  inputString: string;
  error?: boolean;
}

const ShortId: React.FC<ShortIdProps> = ({inputString, error}) => {
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
        <ButtonLink
          className={styles.copy}
          onClick={handleCopy}
          data-testid={`CommitPath_copy`}
        >
          {`${shortId} `}
          {!copied && (
            <Icon small color="plum">
              <CopySVG className={styles.icon} />
            </Icon>
          )}
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
