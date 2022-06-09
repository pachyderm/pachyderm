import {
  Group,
  Button,
  DefaultDropdown,
  HamburgerSVG,
  DropdownItem,
  SupportSVG,
  EducationSVG,
} from '@pachyderm/components';
import React, {useState} from 'react';

import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';

import Account from './components/Account';
import TutorialsMenu from './components/TutorialsMenu';
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
  const [tutorialsMenuVisible, setTutorialsMenu] = useState(false);
  const [stickTutorialsMenu, setStickTutorialsMenu] = useState(false);
  const [activeTutorial] = useLocalProjectSettings({
    projectId: projectId || 'default',
    key: 'active_tutorial',
  });

  const onDropdownMenuSelect = (id: string) => {
    switch (id) {
      case 'support':
        return window.open('mailto:support@pachyderm.com');
      case 'tutorial':
        return setTutorialsMenu(true);
      default:
        return null;
    }
  };

  let menuItems: DropdownItem[] = [
    {
      id: 'tutorial',
      content: 'Learn Pachyderm',
      closeOnClick: true,
      buttonStyle: 'tertiary',
      IconSVG: EducationSVG,
    },
  ];

  if (showSupport) {
    menuItems = [
      ...menuItems,
      {
        id: 'support',
        content: 'Contact Support',
        closeOnClick: true,
        buttonStyle: 'tertiary',
        IconSVG: SupportSVG,
      },
    ];
  }

  return (
    <>
      <Group spacing={16} align="center" className={styles.responsiveHide}>
        {!activeTutorial && (
          <Button
            onClick={() => setTutorialsMenu(true)}
            buttonType="tertiary"
            IconSVG={EducationSVG}
            iconPosition="start"
          >
            Learn Pachyderm
          </Button>
        )}
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
      {tutorialsMenuVisible && (
        <TutorialsMenu
          projectId={projectId}
          setTutorialsMenu={setTutorialsMenu}
          stickTutorialsMenu={stickTutorialsMenu}
          setStickTutorialsMenu={setStickTutorialsMenu}
        />
      )}
    </>
  );
};

export default HeaderButtons;
