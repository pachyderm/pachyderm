import {SkeletonDisplayText} from '@pachyderm/components';
import classnames from 'classnames';
import React from 'react';

import styles from './JobListSkeleton.module.css';

interface JobListSkeletonProps {
  expandActions?: boolean;
}

const JobListSkeleton: React.FC<JobListSkeletonProps> = ({
  expandActions = false,
}) => {
  return (
    <ul
      aria-busy="true"
      aria-live="polite"
      data-testid="JobListSkeleton__list"
      className={classnames(styles.base, {
        [styles.expandActions]: expandActions,
      })}
    >
      <li className={styles.item}>
        <SkeletonDisplayText className={styles.skeleton} />
        <SkeletonDisplayText className={styles.skeleton} />
      </li>
      <li className={styles.item}>
        <SkeletonDisplayText className={styles.skeleton} />
        <SkeletonDisplayText className={styles.skeleton} />
      </li>
      <li className={styles.item}>
        <SkeletonDisplayText className={styles.skeleton} />
        <SkeletonDisplayText className={styles.skeleton} />
      </li>
      <li className={styles.item}>
        <SkeletonDisplayText className={styles.skeleton} />
        <SkeletonDisplayText className={styles.skeleton} />
      </li>
      <li className={styles.item}>
        <SkeletonDisplayText className={styles.skeleton} />
        <SkeletonDisplayText className={styles.skeleton} />
      </li>
      <li className={styles.item}>
        <SkeletonDisplayText className={styles.skeleton} />
        <SkeletonDisplayText className={styles.skeleton} />
      </li>
      <li className={styles.item}>
        <SkeletonDisplayText className={styles.skeleton} />
        <SkeletonDisplayText className={styles.skeleton} />
      </li>
      <li className={styles.item}>
        <SkeletonDisplayText className={styles.skeleton} />
        <SkeletonDisplayText className={styles.skeleton} />
      </li>
      <li className={styles.item}>
        <SkeletonDisplayText className={styles.skeleton} />
        <SkeletonDisplayText className={styles.skeleton} />
      </li>
    </ul>
  );
};

export default JobListSkeleton;
