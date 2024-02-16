import React from 'react';

import Header from '@dash-frontend/components/Header';
import HeaderDropdown from '@dash-frontend/components/HeaderDropdown';
import {useEnterpriseActive} from '@dash-frontend/hooks/useEnterpriseActive';
import {Group, LogoElephant, LogoHpe} from '@pachyderm/components';

import styles from './LandingHeader.module.css';

type LandingHeaderProps = {
  disableBranding?: boolean;
};

export const LandingHeader: React.FC<LandingHeaderProps> = ({
  disableBranding = false,
}) => {
  const {enterpriseActive, loading} = useEnterpriseActive(disableBranding);

  const logo = enterpriseActive ? (
    <LogoHpe aria-describedby="logo-title" />
  ) : (
    <LogoElephant />
  );

  return (
    <Header>
      <Group justify="stretch" align="center">
        <Group align="center" justify="center" spacing={24}>
          {!loading && (
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
                    <span className={styles.hpeLogoSpan}>
                      ML Data Management
                    </span>
                  </>
                ) : (
                  'Console'
                )}
              </h5>
            </a>
          )}
        </Group>
        <HeaderDropdown errorPage={disableBranding} />
      </Group>
    </Header>
  );
};

export const StaticLandingHeader: React.FC = () => {
  return (
    <Header>
      <Group justify="stretch" align="center">
        <Group align="center" justify="center" spacing={24}>
          <a className={styles.logoLink} href="/">
            <h5 className={styles.pachydermLogoHeader}>Console</h5>
          </a>
        </Group>
      </Group>
    </Header>
  );
};
