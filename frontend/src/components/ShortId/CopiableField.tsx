import classnames from 'classnames';
import React, {useCallback} from 'react';

import {
  ButtonLink,
  CopySVG,
  Icon,
  SuccessCheckmark,
  Tooltip,
  useClipboardCopy,
} from '@pachyderm/components';

import styles from './ShortId.module.css';

interface ShortIdProps {
  children?: React.ReactNode;
  inputString: string;
  successCheckmarkAriaLabel?: string;
  error?: boolean;
  inline?: boolean;
}

const CopiableField: React.FC<ShortIdProps> = ({
  inputString,
  children,
  successCheckmarkAriaLabel,
  inline = false,
}) => {
  const {copy, copied, reset} = useClipboardCopy(inputString);

  const handleCopy = useCallback(() => {
    copy();
    setTimeout(reset, 1000);
  }, [copy, reset]);

  return (
    <Tooltip tooltipText={`Click to copy: ${inputString}`}>
      <ButtonLink
        className={classnames(styles.copy, {[styles.inline]: inline})}
        onClick={handleCopy}
      >
        {children}
        {!copied && (
          <Icon small color="plum">
            <CopySVG className={styles.icon} />
          </Icon>
        )}
        <SuccessCheckmark
          show={copied}
          className={styles.icon}
          aria-label={successCheckmarkAriaLabel}
        />
      </ButtonLink>
    </Tooltip>
  );
};

export default CopiableField;
