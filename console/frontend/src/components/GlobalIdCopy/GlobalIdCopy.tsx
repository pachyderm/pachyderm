import React, {useCallback, useState} from 'react';

import {
  CopySVG,
  useClipboardCopy,
  SuccessCheckmark,
  Icon,
  Button,
  IdText,
} from '@pachyderm/components';

import styles from './GlobalIdCopy.module.css';

interface GlobalIdCopyProps {
  id: string;
  shortenId?: boolean;
}

const GlobalIdCopy: React.FC<GlobalIdCopyProps> = ({id, shortenId = false}) => {
  const {copy, copied, reset} = useClipboardCopy(id);
  const [showSuccessIcon, setShowSuccessIcon] = useState(false);

  const handleCopy = useCallback(() => {
    copy();
  }, [copy]);

  return (
    <span
      className={styles.base}
      data-testid="GlobalIdCopy__id"
      onClick={handleCopy}
      onMouseOver={() => setShowSuccessIcon(true)}
      onMouseLeave={() => {
        setShowSuccessIcon(false);
        reset();
      }}
    >
      {shortenId ? `${id.slice(0, 6)}...` : <IdText>{id}</IdText>}
      {showSuccessIcon && copied ? (
        <Icon small className={styles.copyCheckmark}>
          <SuccessCheckmark show={true} aria-label="ID copied successfully" />
        </Icon>
      ) : (
        <Button
          buttonType="ghost"
          className={styles.copyButton}
          onClick={handleCopy}
          aria-label="Copy ID"
          IconSVG={CopySVG}
        />
      )}
    </span>
  );
};

export default GlobalIdCopy;
