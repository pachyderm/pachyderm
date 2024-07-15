import classNames from 'classnames';
import React, {HTMLAttributes} from 'react';

import {PureCheckbox} from '@pachyderm/components';

import styles from './Head.module.css';

export interface HeadProps extends HTMLAttributes<HTMLTableSectionElement> {
  sticky?: boolean;
  screenReaderOnly?: boolean;
  hasCheckbox?: boolean;
  isSelected?: boolean;
  onClick?: () => void;
}

const Head: React.FC<HeadProps> = ({
  children,
  className,
  sticky,
  screenReaderOnly: _screenReaderOnly = false,
  isSelected = false,
  hasCheckbox = false,
  onClick,
  ...rest
}) => {
  const classes = classNames(styles.base, className, {
    [styles.sticky]: sticky,
    [styles.hasCheckbox]: Boolean(hasCheckbox),
  });

  return (
    <thead {...rest} className={classes}>
      {hasCheckbox && onClick && (
        <tr>
          <th>
            <PureCheckbox
              className={styles.checkbox}
              selected={isSelected}
              onChange={onClick}
            />
          </th>
        </tr>
      )}
      {children}
    </thead>
  );
};

export default Head;
