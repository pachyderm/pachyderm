import {Project} from '@graphqlTypes';
import React from 'react';

import Header from '@dash-frontend/components/Header';
import HeaderDropdown from '@dash-frontend/components/HeaderDropdown';
import {useEnterpriseActive} from '@dash-frontend/hooks/useEnterpriseActive';
import {Group, LogoElephant, LogoHpe} from '@pachyderm/components';

import styles from './LandingHeader.module.css';

type LandingHeaderProps = {
  projects?: Project[];
  disableBranding?: boolean;
};

const LandingHeader: React.FC<LandingHeaderProps> = ({
  projects = [],
  disableBranding = false,
}) => {
  const {enterpriseActive} = useEnterpriseActive(disableBranding);

  const logo = enterpriseActive ? (
    <LogoHpe aria-describedby="logo-title" />
  ) : (
    <LogoElephant />
  );

  return (
    <Header>
      <Group justify="stretch" align="center">
        <Group align="center" justify="center" spacing={24}>
          <a className={styles.logoLink} href="/">
            {!disableBranding && logo}
            <h5
              className={
                enterpriseActive
                  ? styles.hpeLogoHeader
                  : styles.pachydermLogoHeader
              }
            >
              {enterpriseActive ? (
                <>
                  HPE{' '}
                  <span className={styles.hpeLogoSpan}>ML Data Management</span>
                </>
              ) : (
                'Console'
              )}
            </h5>
          </a>
        </Group>
        <HeaderDropdown errorPage={disableBranding} />
      </Group>
    </Header>
  );
};

export default LandingHeader;
