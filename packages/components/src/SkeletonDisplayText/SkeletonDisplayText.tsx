import classnames from 'classnames';
import React, {memo} from 'react';

import styles from './SkeletonDisplayText.module.css';

type SkeletonDisplayTextProps = {
  blueShimmer?: boolean;
};

const SkeletonDisplayText: React.FC<SkeletonDisplayTextProps> = ({
  blueShimmer = false,
  ...rest
}) => (
  <div
    className={classnames(styles.base, {[styles.blueShimmer]: blueShimmer})}
    {...rest}
  />
);

export default memo(SkeletonDisplayText);
