import {SkeletonDisplayText} from '@pachyderm/components';
import React from 'react';

import useAccount from '@dash-frontend/hooks/useAccount';
import useLoggedIn from '@dash-frontend/hooks/useLoggedIn';

import styles from './Account.module.css';

const Account: React.FC = () => {
  const loggedIn = useLoggedIn();
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
