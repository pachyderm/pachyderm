import React from 'react';

import Tooltip from './Tooltip';

export default {title: 'Tooltip'};

export const Default = () => {
  return (
    <Tooltip tooltipText="A default tooltip">
      <span>Show Tooltip</span>
    </Tooltip>
  );
};

export const Large = () => {
  const tooltipText = `This is a tooltip with lots of text and a forced bottom placement. It has 1rem of padding and is left aligned.
Additionally its max width is 18.75rem.`;
  return (
    <Tooltip tooltipText={tooltipText} allowedPlacements={['bottom']}>
      <span>Show Tooltip</span>
    </Tooltip>
  );
};
