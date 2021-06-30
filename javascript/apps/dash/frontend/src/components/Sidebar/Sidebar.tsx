import {ButtonLink, ExitSVG} from '@pachyderm/components';
import classnames from 'classnames';
import React, {HTMLAttributes, useEffect, useState} from 'react';

import {SidebarSize} from '@dash-frontend/lib/types';

import styles from './Sidebar.module.css';

interface SidebarProps extends HTMLAttributes<HTMLDivElement> {
  overlay?: boolean;
  fixed?: boolean;
  onClose?: () => void;
  size?: SidebarSize;
}

const Sidebar: React.FC<SidebarProps> = ({
  children,
  overlay = false,
  onClose,
  className,
  fixed = false,
  size = 'sm',
  ...rest
}) => {
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    setIsOpen(true);
  }, []);

  return (
    <>
      {overlay && (
        <div
          className={classnames(styles.overlay, {
            [styles.open]: isOpen,
            [styles[size]]: true,
          })}
        />
      )}

      <div
        className={classnames(
          styles.base,
          {
            [styles.fixed]: fixed,
            [styles.open]: isOpen,
            [styles[size]]: true,
          },
          className,
        )}
        {...rest}
      >
        {onClose && (
          <div className={styles.closeContainer}>
            <ButtonLink className={styles.closeButton} onClick={onClose}>
              <ExitSVG aria-label="Close" className={styles.closeSvg} />
            </ButtonLink>
          </div>
        )}

        {children}
      </div>
    </>
  );
};

export default Sidebar;
