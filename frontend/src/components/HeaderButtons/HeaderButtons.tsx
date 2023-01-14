import React, {useState} from 'react';

import {
  Group,
  Button,
  DefaultDropdown,
  HamburgerSVG,
  DropdownItem,
  SupportSVG,
} from '@pachyderm/components';

import Account from './components/Account';
import styles from './HeaderButtons.module.css';

type HeaderButtonsProps = {
  showSupport?: boolean;
  showAccount?: boolean;
};

const HeaderButtons: React.FC<HeaderButtonsProps> = ({
  showSupport = false,
  showAccount = false,
}) => {
  /* Tutorial is temporarily disabled because of "Project" Console Support */

  const onDropdownMenuSelect = (id: string) => {
    switch (id) {
      case 'support':
        return window.open('mailto:support@pachyderm.com');
      /* Tutorial is temporarily disabled because of "Project" Console Support */
      default:
        return null;
    }
  };

  let menuItems: DropdownItem[] = [
    /* Tutorial is temporarily disabled because of "Project" Console Support */
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
        {/* Tutorial is temporarily disabled because of "Project" Console Support */}
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
      {/* Tutorial is temporarily disabled because of "Project" Console Support */}
    </>
  );
};

export default HeaderButtons;
