import {
  Group,
  Button,
  DefaultDropdown,
  HamburgerSVG,
  DropdownItem,
  SupportSVG,
} from '@pachyderm/components';
import React, {Children} from 'react';

import RunTutorialButton from '@dash-frontend/components/RunTutorialButton';
import useRunTutorialButton from '@dash-frontend/components/RunTutorialButton/hooks/useRunTutorialButton';

import styles from './HeaderButtons.module.css';

type HeaderButtonsProps = {
  projectId?: string;
  showSupport?: boolean;
};

const HeaderButtons: React.FC<HeaderButtonsProps> = ({
  projectId,
  showSupport = false,
  children,
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
        {Children.count(children) !== 0 && <div className={styles.divider} />}
        {children}
      </Group>
      <div className={styles.responsiveShow}>
        <DefaultDropdown
          items={menuItems}
          className={styles.dropdown}
          onSelect={onDropdownMenuSelect}
          buttonOpts={{hideChevron: true}}
          menuOpts={{pin: 'right'}}
        >
          <HamburgerSVG />
        </DefaultDropdown>
      </div>
    </>
  );
};

export default HeaderButtons;
