import {SkeletonBodyText} from '@pachyderm/components';
import React, {HTMLAttributes} from 'react';

import styles from './Description.module.css';

interface DescriptionProps extends HTMLAttributes<HTMLElement> {
  term: string;
  loading?: boolean;
  lines?: number;
}

const Description: React.FC<DescriptionProps> = ({
  term,
  children,
  loading = false,
  lines = 1,
  ...rest
}) => {
  return (
    <>
      <dt className={styles.term}>{term}</dt>
      <dd className={styles.description} {...rest}>
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
