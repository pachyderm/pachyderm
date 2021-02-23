import classnames from 'classnames';
import range from 'lodash/range';
import React, {memo} from 'react';

import styles from './SkeletonBodyText.module.css';

interface SkeletonBodyText {
  lines?: number;
  disableShimmer?: boolean;
}

const SkeletonBodyText: React.FC<SkeletonBodyText> = ({
  lines = 1,
  disableShimmer = false,
  ...props
}) => (
  <div
    className={classnames(styles.base, {
      [styles.disableShimmer]: disableShimmer,
    })}
    {...props}
  >
    {range(lines).map((i) => (
      <div className={styles.skeletonBodyText} key={i} />
    ))}
  </div>
);

export default memo(SkeletonBodyText);
