import {Group, ButtonLink} from '@pachyderm/components';
import React from 'react';

import Account from './components/Account';
import styles from './LandingHeader.module.css';
import {ReactComponent as LogoElephant} from './LogoElephant.svg';

const LandingHeader = () => (
  <Group justify="stretch" align="center">
    <Group align="center" justify="center" spacing={24}>
      <a className={styles.logo} href="/">
        <LogoElephant />
        <span className={styles.dashboard}>Dashboard</span>
      </a>
      <div className={styles.divider} />
      <span className={styles.workspaceName}>
        Workspace &lt;Elegant Elephant&gt;
      </span>
    </Group>

    <Group spacing={24} align="center">
      <ButtonLink className={styles.support} small>
        Support
      </ButtonLink>
      <div className={styles.divider} />
      <Account />
    </Group>
  </Group>
);

export default LandingHeader;
