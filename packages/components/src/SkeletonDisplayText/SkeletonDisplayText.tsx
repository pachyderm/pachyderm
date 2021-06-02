import classnames from 'classnames';
import React, {HTMLAttributes} from 'react';

import styles from './SkeletonDisplayText.module.css';

interface SkeletonDisplayTextProps extends HTMLAttributes<HTMLDivElement> {
  color?: 'blue' | 'grey';
}

const SkeletonDisplayText: React.FC<SkeletonDisplayTextProps> = ({
  color = '',
  className,
  ...rest
}) => (
  <div
    className={classnames(
      styles.base,
      {...(Boolean(color) && {[styles[color]]: true})},
      className,
    )}
    {...rest}
  />
);

export default SkeletonDisplayText;
