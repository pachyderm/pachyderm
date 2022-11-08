import {Project} from '@graphqlTypes';
import React from 'react';

import Header from '@dash-frontend/components/Header';
import HeaderButtons from '@dash-frontend/components/HeaderButtons';
import {Group} from '@pachyderm/components';

import styles from './LandingHeader.module.css';
import {ReactComponent as LogoElephant} from './LogoElephant.svg';

type LandingHeaderProps = {
  projects?: Project[];
};

const LandingHeader: React.FC<LandingHeaderProps> = ({projects = []}) => {
  return (
    <Header>
      <Group justify="stretch" align="center">
        <Group align="center" justify="center" spacing={24}>
          <a className={styles.logo} href="/">
            <LogoElephant />
            <h5 className={styles.dashboard}>Console</h5>
          </a>
        </Group>
        <HeaderButtons
          showSupport
          showAccount
          projectId={projects.length > 0 ? projects[0].id : undefined}
        />
      </Group>
    </Header>
  );
};

export default LandingHeader;
