import React from 'react';

import {Group, Icon, BoxSVG} from '@pachyderm/components';

const PageHeading: React.FC = ({children}) => (
  <Group spacing={8} align="center">
    <Icon aria-hidden={true}>
      <BoxSVG />
    </Icon>

    <h2>{children}</h2>
  </Group>
);

export default PageHeading;
