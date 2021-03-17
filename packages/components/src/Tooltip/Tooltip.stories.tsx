import React from 'react';

import Tooltip from './Tooltip';

export default {title: 'Tooltip'};

export const Default = () => {
  return (
    <>
      <Tooltip
        tooltipKey="tooltip"
        tooltipText="A default tooltip"
        placement="bottom"
        defaultShow
      >
        <span>Show Default Tooltip</span>
      </Tooltip>
    </>
  );
};

export const Large = () => {
  const tooltipText = `This is a large tooltip, it is used when 
the tooltip contains a large amount of text. It's padding is 1rem and it's text is left aligned. 
Additionally its max width is 18.75rem.`;
  return (
    <>
      <Tooltip
        tooltipKey="tooltip"
        tooltipText={tooltipText}
        size="large"
        placement="bottom"
        defaultShow
      >
        <span>Show Large Tooltip</span>
      </Tooltip>
    </>
  );
};

export const ExtraLarge = () => {
  const tooltipText = `This is an extra large tooltip, it is used when 
the tooltip contains a really large amount of text. It's padding is 1.25rem and it's text is left aligned. 
Additionally its max width is 24.75rem. It's padding is 1.25rem and it's text is left aligned. 
Additionally its max width is 24.75rem`;
  return (
    <>
      <Tooltip
        tooltipKey="tooltip"
        tooltipText={tooltipText}
        size="extraLarge"
        placement="bottom"
        defaultShow
      >
        <span>Show Extra Large Tooltip</span>
      </Tooltip>
    </>
  );
};
