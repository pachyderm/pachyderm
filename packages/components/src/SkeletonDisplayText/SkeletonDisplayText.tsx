import classnames from 'classnames';
import React, {HTMLAttributes} from 'react';

import styles from './SkeletonDisplayText.module.css';

interface SkeletonDisplayTextProps extends HTMLAttributes<HTMLDivElement> {
  blueShimmer?: boolean;
}

const SkeletonDisplayText: React.FC<SkeletonDisplayTextProps> = ({
  blueShimmer = false,
  className,
  ...rest
}) => (
  <div
    className={classnames(
      styles.base,
      {[styles.blueShimmer]: blueShimmer},
      className,
    )}
    {...rest}
  />
);

export default SkeletonDisplayText;
