import {EnterpriseState, Project} from '@graphqlTypes';
import React from 'react';

import Header from '@dash-frontend/components/Header';
import HeaderButtons from '@dash-frontend/components/HeaderButtons';
import {useGetEnterpriseInfoQuery} from '@dash-frontend/generated/hooks';
import {Group, LogoElephant, LogoHpe} from '@pachyderm/components';

import styles from './LandingHeader.module.css';

type LandingHeaderProps = {
  projects?: Project[];
};

const LandingHeader: React.FC<LandingHeaderProps> = ({projects = []}) => {
  const {data} = useGetEnterpriseInfoQuery();
  const enterpriseActive =
    data?.enterpriseInfo.state === EnterpriseState.ACTIVE;

  return (
    <Header>
      <Group justify="stretch" align="center">
        <Group align="center" justify="center" spacing={24}>
          <a className={styles.logoLink} href="/">
            {enterpriseActive ? (
              <LogoHpe aria-describedby="logo-title" />
            ) : (
              <LogoElephant />
            )}
            <h5
              className={
                enterpriseActive
                  ? styles.hpeLogoHeader
                  : styles.pachydermLogoHeader
              }
            >
              {enterpriseActive ? (
                <>
                  HPE <span className={styles.hpeLogoSpan}>MLDM</span>
                </>
              ) : (
                'Console'
              )}
            </h5>
          </a>
        </Group>
        <HeaderButtons showSupport showAccount />
      </Group>
    </Header>
  );
};

export default LandingHeader;
