import classnames from 'classnames';
import noop from 'lodash/noop';
import React, {useCallback} from 'react';

import {ButtonLink} from '../../ButtonLink';
import useClipboardCopy from '../../hooks/useClipboardCopy';
import {Icon} from '../../Icon';
import {useNotificationBanner} from '../../NotificationBanner';
import {CopySVG} from '../../Svg';
import {CodeTextBlock} from '../../Text';

import styles from './Description.module.css';

interface DescriptionProps {
  children?: React.ReactNode;
  action?: React.ReactNode;
  disabled?: boolean;
  hideCopy?: boolean;
  id: string;
  copyText?: string;
  title: string;
  className?: string;
  noteText?: string;
  asListItem?: boolean;
}

interface DescriptionWrapperProps {
  children?: React.ReactNode;
  className?: string;
  asListItem: boolean;
}

const DescriptionWrapper: React.FC<DescriptionWrapperProps> = ({
  asListItem,
  className,
  children,
}) => {
  return asListItem ? (
    <li className={classnames(styles.base, className)}>{children}</li>
  ) : (
    <div className={classnames(styles.base, className)}>{children}</div>
  );
};

const Description: React.FC<DescriptionProps> = ({
  disabled = false,
  copyText,
  children,
  action,
  title,
  className,
  noteText,
  asListItem = true,
}) => {
  const {copy, supported} = useClipboardCopy(copyText || '');
  const {add} = useNotificationBanner();

  const handleClick = useCallback(() => {
    copy();
    add('Command Line Copied');
  }, [add, copy]);

  return (
    <DescriptionWrapper className={className} asListItem={asListItem}>
      <div className={styles.title}>{title}</div>

      {((copyText && supported) || action) && (
        <div className={styles.copy}>{action}</div>
      )}

      {children && (
        <div className={styles.mono}>
          <CodeTextBlock
            role="button"
            onClick={!disabled && copyText ? handleClick : noop}
            aria-disabled={disabled}
            className={styles.mono}
            tabIndex={0}
          >
            {children}
          </CodeTextBlock>

          {!disabled && (
            <ButtonLink
              data-testid="Description__copy"
              onClick={handleClick}
              className={styles.copyIcon}
              aria-label="Copy"
            >
              <Icon color="plum">
                <CopySVG />
              </Icon>
            </ButtonLink>
          )}
        </div>
      )}
      {noteText && <p className={styles.note}>Note: {noteText}</p>}
    </DescriptionWrapper>
  );
};

export default Description;
