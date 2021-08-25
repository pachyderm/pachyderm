import {Group, ButtonLink, Link} from '@pachyderm/components';
import React, {useState} from 'react';

import ConnectModal from '@dash-frontend/components/ConnectModal';
import Header from '@dash-frontend/components/Header';
import {useWorkspace} from '@dash-frontend/hooks/useWorkspace';

import Account from './components/Account';
import styles from './LandingHeader.module.css';
import {ReactComponent as LogoElephant} from './LogoElephant.svg';

const LandingHeader = () => {
  const {workspaceName, pachdAddress, pachVersion, hasConnectInfo} =
    useWorkspace();
  const [connectModalShow, showConnectModal] = useState(false);

  return (
    <Header>
      <Group justify="stretch" align="center">
        <Group align="center" justify="center" spacing={24}>
          <a className={styles.logo} href="/">
            <LogoElephant />
            <span className={styles.dashboard}>Console</span>
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
          {hasConnectInfo && (
            <ButtonLink
              className={styles.support}
              onClick={() => showConnectModal(true)}
            >
              Connect to Workspace
            </ButtonLink>
          )}
          <Link
            className={styles.support}
            small
            href="mailto:support@pachyderm.com"
          >
            Support
          </Link>
          <div className={styles.divider} />
          <Account />
        </Group>
      </Group>
      {hasConnectInfo && (
        <ConnectModal
          show={connectModalShow}
          onHide={() => showConnectModal(false)}
          workspaceName={workspaceName || ''}
          pachdAddress={pachdAddress || ''}
          pachVersion={pachVersion || ''}
        />
      )}
    </Header>
  );
};

export default LandingHeader;
