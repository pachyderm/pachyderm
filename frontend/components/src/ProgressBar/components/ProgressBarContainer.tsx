import React from 'react';

import {useProgressBar} from '@pachyderm/components/ProgressBar';

import {Group} from '../../Group';

const ProgressBarContainer: React.FC = ({children}) => {
  const {isVertical} = useProgressBar();

  return (
    <Group justify="center" align="center" vertical={isVertical}>
      {children}
    </Group>
  );
};

export default ProgressBarContainer;
