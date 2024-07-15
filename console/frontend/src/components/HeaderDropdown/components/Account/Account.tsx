import React from 'react';

import {useAccount} from '@dash-frontend/hooks/useAccount';
import useLoggedIn from '@dash-frontend/hooks/useLoggedIn';

import styles from './Account.module.css';

const Account: React.FC = () => {
  const {loggedIn} = useLoggedIn();
  const {displayName} = useAccount(loggedIn);

  if (!loggedIn) {
    return null;
  }

  return (
    <div className={styles.base}>
      <div className={styles.letter}>
        <h5>{displayName ? displayName[0] : 'U'}</h5>
      </div>
      <h5>{displayName}</h5>
    </div>
  );
};

export default Account;
