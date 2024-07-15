import React from 'react';

import {EMAIL_SUPPORT, SLACK_SUPPORT} from '@dash-frontend/constants/links';
import {useEnterpriseActive} from '@dash-frontend/hooks/useEnterpriseActive';
import {useVersion} from '@dash-frontend/hooks/useVersion';
import {getReleaseVersion} from '@dash-frontend/lib/runtimeVariables';
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
  const {version} = useVersion({enabled: !errorPage});
  const consoleVersion = getReleaseVersion();

  let pachdVersionString = '';

  if (version?.major === 0 && version?.minor === 0 && version?.micro === 0) {
    pachdVersionString = version.gitCommit || 'unknown';
  } else {
    pachdVersionString = version
      ? `${version.major}.${version.minor}.${version.micro}${version.additional}`
      : 'unknown';
  }
  const emailLinkWithPrefill = `${EMAIL_SUPPORT}?subject=Console%20Support&body=Console%20version:%20${consoleVersion}%0APachd%20version:%20${pachdVersionString}`;

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
          {consoleVersion && <div>Console {consoleVersion}</div>}
          {version && <div>Pachd {pachdVersionString}</div>}
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
