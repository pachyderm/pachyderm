import React from 'react';

import {EMAIL_SUPPORT, SLACK_SUPPORT} from '@dash-frontend/constants/links';
import {useGetVersionInfoQuery} from '@dash-frontend/generated/hooks';
import {useEnterpriseActive} from '@dash-frontend/hooks/useEnterpriseActive';
import {
  CaptionTextSmall,
  HamburgerSVG,
  SupportSVG,
  Dropdown,
} from '@pachyderm/components';

import Account from './components/Account';
import styles from './HeaderDropdown.module.css';

type HeaderDropdownProps = {
  errorPage?: boolean;
};

const HeaderDropdown: React.FC<HeaderDropdownProps> = ({errorPage}) => {
  const {enterpriseActive} = useEnterpriseActive(errorPage);
  const {data: version} = useGetVersionInfoQuery({skip: errorPage});

  const pachdVersion = version?.versionInfo.pachdVersion;
  const pachdVersionString = pachdVersion
    ? `${pachdVersion.major}.${pachdVersion.minor}.${pachdVersion.micro}${pachdVersion.additional}`
    : 'unknown';
  const emailLinkWithPrefill = `${EMAIL_SUPPORT}?subject=Console%20Support&body=Console%20version:%20${version?.versionInfo.consoleVersion}%0APachd%20version:%20${pachdVersionString}`;

  const onDropdownMenuSelect = (id: string) => {
    switch (id) {
      case 'support':
        return window.open(
          enterpriseActive ? emailLinkWithPrefill : SLACK_SUPPORT,
        );
      default:
        return null;
    }
  };

  return (
    <Dropdown onSelect={onDropdownMenuSelect}>
      <Dropdown.Button
        aria-label="header menu"
        hideChevron
        buttonType="tertiary"
        IconSVG={HamburgerSVG}
      />
      <Dropdown.Menu pin="right" className={styles.menu}>
        <Account />
        <div className={styles.versions}>
          {version?.versionInfo.consoleVersion && (
            <div>Console {version?.versionInfo.consoleVersion}</div>
          )}
          {pachdVersion && <div>Pachd {pachdVersionString}</div>}
        </div>
        <Dropdown.MenuItem
          id="support"
          closeOnClick
          buttonStyle="tertiary"
          IconSVG={SupportSVG}
          aria-label={enterpriseActive ? 'email support' : 'open slack support'}
        >
          Contact Support
        </Dropdown.MenuItem>
        <CaptionTextSmall className={styles.copyright} color="white">
          © {new Date().getFullYear()}, HPE
        </CaptionTextSmall>
      </Dropdown.Menu>
    </Dropdown>
  );
};

export default HeaderDropdown;
