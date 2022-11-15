import classNames from 'classnames';
import React from 'react';

import styles from './TabsHeader.module.css';

const TabsHeader: React.FC<React.HTMLAttributes<HTMLUListElement>> = ({
  children,
  className,
  ...rest
}) => {
  const mergedClasses = classNames(styles.base, className);

  return (
    <ul role="tablist" className={mergedClasses} {...rest}>
      {children}
    </ul>
  );
};

export default TabsHeader;
