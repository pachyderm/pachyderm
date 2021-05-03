import {Group, ButtonLink} from '@pachyderm/components';
import React from 'react';

import Header from '@dash-frontend/components/Header';
import {useWorkspace} from '@dash-frontend/hooks/useWorkspace';

import Account from './components/Account';
import styles from './LandingHeader.module.css';
import {ReactComponent as LogoElephant} from './LogoElephant.svg';

const LandingHeader = () => {
  const {workspaceName} = useWorkspace();

  return (
    <Header>
      <Group justify="stretch" align="center">
        <Group align="center" justify="center" spacing={24}>
          <a className={styles.logo} href="/">
            <LogoElephant />
            <span className={styles.dashboard}>Dashboard</span>
          </a>
          {workspaceName && (
            <>
              <div className={styles.divider} />
              <span className={styles.workspaceName}>
                Workspace {workspaceName}
              </span>
            </>
          )}
        </Group>

        <Group spacing={24} align="center">
          <ButtonLink className={styles.support} small>
            Support
          </ButtonLink>
          <div className={styles.divider} />
          <Account />
        </Group>
      </Group>
    </Header>
  );
};

export default LandingHeader;
