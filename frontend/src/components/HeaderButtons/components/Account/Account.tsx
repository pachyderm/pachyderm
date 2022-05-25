import React from 'react';

import useAccount from '@dash-frontend/hooks/useAccount';
import useLoggedIn from '@dash-frontend/hooks/useLoggedIn';

import styles from './Account.module.css';

const Account: React.FC = () => {
  const {loggedIn} = useLoggedIn();
  const {displayName, loading} = useAccount({skip: !loggedIn});

  if (loading) {
    return (
      <div className={styles.loaderContainer} data-testid="Account__loader" />
    );
  }

  if (!loggedIn) {
    return null;
  }

  return <strong className={styles.base}>Hello, {displayName}!</strong>;
};

export default Account;
