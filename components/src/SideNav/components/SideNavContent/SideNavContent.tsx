import classnames from 'classnames';
import React, {HTMLAttributes} from 'react';

import useSideNav from '../../hooks/useSideNav';

import styles from './SideNavContent.module.css';

const SideNavContent: React.FC<HTMLAttributes<HTMLDivElement>> = ({
  children,
  className,
  ...rest
}) => {
  const {minimized} = useSideNav();

  const mergedClassNames = classnames(
    styles.base,
    {[styles.minimized]: minimized},
    className,
  );

  return (
    <div
      role={minimized ? 'button' : 'presentation'}
      tabIndex={minimized ? 0 : -1}
      className={mergedClassNames}
      {...rest}
    >
      {children}
    </div>
  );
};

export default SideNavContent;
