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
            <h5 className={styles.dashboard}>Console</h5>
          </a>
          {workspaceName && (
            <>
              <div className={styles.divider} />

              {hasConnectInfo ? (
                <ButtonLink
                  className={styles.support}
                  onClick={() => showConnectModal(true)}
                >
                  <h6 className={styles.support}>
                    Connect to Workspace {workspaceName}
                  </h6>
                </ButtonLink>
              ) : (
                <h6 className={styles.workspaceName}>
                  Workspace {workspaceName}
                </h6>
              )}
            </>
          )}
        </Group>

        <Group spacing={24} align="center">
          <Link className={styles.support} href="mailto:support@pachyderm.com">
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
