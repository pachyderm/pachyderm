import {SkeletonDisplayText} from '@pachyderm/components';
import React from 'react';

import useAccount from '@dash-frontend/hooks/useAccount';
import useAuth from '@dash-frontend/hooks/useAuth';

import styles from './Account.module.css';

const Account: React.FC = () => {
  const {loggedIn} = useAuth();
  const {displayName, loading} = useAccount({skip: !loggedIn});

  if (loading || !loggedIn) {
    return (
      <div className={styles.loaderContainer} data-testid={'Account__loader'}>
        <SkeletonDisplayText blueShimmer className={styles.loader} />
      </div>
    );
  }

  return <span className={styles.base}>Hello, {displayName}!</span>;
};

export default Account;
