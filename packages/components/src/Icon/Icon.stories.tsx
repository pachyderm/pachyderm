import React from 'react';

import {TimesSVG} from '../Svg';

import {Icon} from './';

/* eslint-disable-next-line import/no-anonymous-default-export */
export default {title: 'Icon'};

export const Default = () => {
  return (
    <Icon color="black">
      <TimesSVG />
    </Icon>
  );
};
