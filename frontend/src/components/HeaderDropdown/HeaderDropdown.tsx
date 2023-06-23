import React from 'react';
import {Route} from 'react-router';

import ActiveProjectModal from '@dash-frontend/components/ActiveProjectModal';
import {EMAIL_SUPPORT, SLACK_SUPPORT} from '@dash-frontend/constants/links';
import {useGetVersionInfoQuery} from '@dash-frontend/generated/hooks';
import {useEnterpriseActive} from '@dash-frontend/hooks/useEnterpriseActive';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  PROJECT_PATH,
  LINEAGE_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';
import {
  CaptionTextSmall,
  HamburgerSVG,
  SupportSVG,
  TerminalSVG,
  Dropdown,
  useModal,
} from '@pachyderm/components';

import Account from './components/Account';
import styles from './HeaderDropdown.module.css';

type HeaderDropdownProps = {
  errorPage?: boolean;
};

const HeaderDropdown: React.FC<HeaderDropdownProps> = ({errorPage}) => {
  const {projectId} = useUrlState();
  const {enterpriseActive} = useEnterpriseActive(errorPage);
  const {data: version} = useGetVersionInfoQuery({skip: errorPage});
  const {
    openModal: openActiveProjectModal,
    closeModal: closeActiveProjectModal,
    isOpen: activeProjectModalIsOpen,
  } = useModal(false);

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
      case 'set-active-project':
        openActiveProjectModal();
        return null;
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
        <Route path={[PROJECT_PATH, LINEAGE_PATH]}>
          <Dropdown.MenuItem
            id="set-active-project"
            closeOnClick
            buttonStyle="tertiary"
            IconSVG={TerminalSVG}
          >
            Set Active Project
          </Dropdown.MenuItem>
        </Route>
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

      {activeProjectModalIsOpen && (
        <ActiveProjectModal
          show={activeProjectModalIsOpen}
          onHide={closeActiveProjectModal}
          projectName={projectId}
        />
      )}
    </Dropdown>
  );
};

export default HeaderDropdown;
