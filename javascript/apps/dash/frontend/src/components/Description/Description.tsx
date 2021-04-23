import {SkeletonBodyText} from '@pachyderm/components';
import React from 'react';

import styles from './Description.module.css';

interface DescriptionProps {
  term: string;
  loading?: boolean;
  lines?: number;
}

const Description: React.FC<DescriptionProps> = ({
  term,
  children,
  loading = false,
  lines = 1,
}) => {
  return (
    <>
      <dt className={styles.term}>{term}</dt>
      <dd className={styles.description}>
        {loading ? (
          <div className={lines === 1 ? styles.singleLineLoading : undefined}>
            <SkeletonBodyText
              lines={lines}
              data-testid={`Description__${term}Skeleton`}
            />
          </div>
        ) : (
          children
        )}
      </dd>
    </>
  );
};

export default Description;
