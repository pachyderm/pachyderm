import React from 'react';

import {InfoSVG} from '../Svg';

import {Icon} from './';

export default {title: 'Icon'};

export const Default = () => {
  return (
    <>
      <Icon color="black">
        <InfoSVG />
      </Icon>
      <Icon color="white">
        <InfoSVG />
      </Icon>
      <Icon color="purple">
        <InfoSVG />
      </Icon>
      <Icon color="silver">
        <InfoSVG />
      </Icon>
      <Icon small>
        <InfoSVG />
      </Icon>
    </>
  );
};
