import classNames from 'classnames';
import React from 'react';

import {CaptionTextSmall} from '../../../Text';
import useSideNav from '../../hooks/useSideNav';

import styles from './SideNavList.module.css';

type SideNavListProps = {
  children?: React.ReactNode;
  noPadding?: boolean;
  border?: boolean;
  label?: string;
};

const SideNavList: React.FC<SideNavListProps> = ({
  noPadding = false,
  border = false,
  label,
  children,
}) => {
  const {minimized} = useSideNav();

  return (
    <ul
      className={classNames(styles.base, {
        [styles.noPadding]: noPadding,
        [styles.border]: border,
      })}
    >
      {label && (
        <li>
          <CaptionTextSmall className={styles.label}>
            {!minimized && label}
          </CaptionTextSmall>
        </li>
      )}
      {children}
    </ul>
  );
};

export default SideNavList;
