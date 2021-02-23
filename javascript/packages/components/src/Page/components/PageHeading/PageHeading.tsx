import React from 'react';

import {Group} from 'Group';
import {Icon} from 'Icon';
import {BoxSVG} from 'Svg';

import styles from './PageHeading.module.css';

const PageHeading: React.FC = ({children}) => (
  <Group spacing={8}>
    <Icon aria-hidden={true}>
      <BoxSVG />{' '}
    </Icon>

    <h1 className={styles.pageHeading}>{children}</h1>
  </Group>
);

export default PageHeading;
