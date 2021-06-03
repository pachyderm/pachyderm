import {ButtonLink, CloseSVG} from '@pachyderm/components';
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
  overlay,
  onClose,
  className,
  fixed,
  size = 'sm',
  ...rest
}) => {
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    setIsOpen(true);
  }, []);

  return (
    <div
      className={classnames(
        styles.base,
        {
          [styles.overlay]: overlay,
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
            <CloseSVG aria-label="Close" className={styles.closeSvg} />
          </ButtonLink>
        </div>
      )}

      {children}
    </div>
  );
};

export default Sidebar;
