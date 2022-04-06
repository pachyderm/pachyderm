import React from 'react';

import {Group} from 'Group';
import {Icon} from 'Icon';
import {BoxSVG} from 'Svg';

const PageHeading: React.FC = ({children}) => (
  <Group spacing={8} align="center">
    <Icon aria-hidden={true}>
      <BoxSVG />
    </Icon>

    <h2>{children}</h2>
  </Group>
);

export default PageHeading;
