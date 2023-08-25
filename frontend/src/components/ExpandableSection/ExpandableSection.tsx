import classnames from 'classnames';
import React from 'react';

import styles from './ExpandableSection.module.css';

interface ExpandableSectionProps {
  children?: React.ReactNode;
  maxHeightRem: number;
  expanded: boolean;
}

const ExpandableFilters: React.FC<ExpandableSectionProps> = ({
  children,
  maxHeightRem,
  expanded,
}) => {
  return (
    <div
      className={classnames(styles.base, {
        [styles.expanded]: expanded,
      })}
      style={{height: expanded ? `${maxHeightRem}rem` : '0rem'}}
    >
      {children}
    </div>
  );
};

export default ExpandableFilters;
