import {
  Group,
  Button,
  DefaultDropdown,
  HamburgerSVG,
  DropdownItem,
  SupportSVG,
} from '@pachyderm/components';
import React from 'react';

import RunTutorialButton from '@dash-frontend/components/RunTutorialButton';
import useRunTutorialButton from '@dash-frontend/components/RunTutorialButton/hooks/useRunTutorialButton';

import Account from './components/Account';
import styles from './HeaderButtons.module.css';

type HeaderButtonsProps = {
  projectId?: string;
  showSupport?: boolean;
  showAccount?: boolean;
};

const HeaderButtons: React.FC<HeaderButtonsProps> = ({
  projectId,
  showSupport = false,
  showAccount = false,
}) => {
  const {startTutorial, tutorialProgress} = useRunTutorialButton(projectId);

  const onDropdownMenuSelect = (id: string) => {
    switch (id) {
      case 'support':
        return window.open('mailto:support@pachyderm.com');
      case 'tutorial':
        return startTutorial();
      default:
        return null;
    }
  };

  let menuItems: DropdownItem[] = [
    {
      id: 'tutorial',
      content: tutorialProgress ? 'Resume Tutorial' : 'Run Tutorial',
      closeOnClick: true,
    },
  ];

  if (showSupport) {
    menuItems = [
      {id: 'support', content: 'Contact Support', closeOnClick: true},
      ...menuItems,
    ];
  }

  return (
    <>
      <Group spacing={16} align="center" className={styles.responsiveHide}>
        <RunTutorialButton projectId={projectId} />
        {showSupport && (
          <Button
            buttonType="tertiary"
            IconSVG={SupportSVG}
            href="mailto:support@pachyderm.com"
          >
            Contact Support
          </Button>
        )}
        {showAccount && (
          <>
            <div className={styles.divider} />
            <Account />
          </>
        )}
      </Group>
      <Group spacing={16} align="center" className={styles.responsiveShow}>
        {showAccount && (
          <>
            <Account />
            <div className={styles.divider} />
          </>
        )}
        <DefaultDropdown
          items={menuItems}
          className={styles.dropdown}
          onSelect={onDropdownMenuSelect}
          buttonOpts={{
            hideChevron: true,
            buttonType: 'tertiary',
            IconSVG: HamburgerSVG,
          }}
          menuOpts={{pin: 'right'}}
        />
      </Group>
    </>
  );
};

export default HeaderButtons;
