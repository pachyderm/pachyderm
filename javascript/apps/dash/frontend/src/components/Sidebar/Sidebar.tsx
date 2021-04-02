import {ButtonLink, CloseSVG} from '@pachyderm/components';
import classnames from 'classnames';
import React, {HTMLAttributes} from 'react';

import styles from './Sidebar.module.css';

interface SidebarProps extends HTMLAttributes<HTMLDivElement> {
  overlay?: boolean;
  onClose?: () => void;
}

const Sidebar: React.FC<SidebarProps> = ({children, overlay, onClose}) => (
  <div className={classnames(styles.base, {[styles.overlay]: overlay})}>
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

export default Sidebar;
