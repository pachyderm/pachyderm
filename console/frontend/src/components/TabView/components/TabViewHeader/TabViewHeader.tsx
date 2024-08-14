import React from 'react';

import {Group} from '@pachyderm/components';

import styles from './TabViewHeader.module.css';

type TabViewHeaderProps = {
  children?: React.ReactNode;
  heading: string;
};

const TabViewHeader: React.FC<TabViewHeaderProps> = ({
  heading = '',
  children,
}) => {
  return (
    <Group justify="stretch">
      <h2 className={styles.heading}>{heading}</h2>
      <Group spacing={8}>{children}</Group>
    </Group>
  );
};

export default TabViewHeader;
