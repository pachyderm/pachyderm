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
      <Icon color="plum">
        <InfoSVG />
      </Icon>
      <Icon color="grey">
        <InfoSVG />
      </Icon>
      <Icon color="green">
        <InfoSVG />
      </Icon>
      <Icon color="blue">
        <InfoSVG />
      </Icon>
      <Icon color="red">
        <InfoSVG />
      </Icon>
      <Icon color="yellow">
        <InfoSVG />
      </Icon>
      <Icon color="highlightGreen">
        <InfoSVG />
      </Icon>
      <Icon color="highlightOrange">
        <InfoSVG />
      </Icon>
      <Icon disabled>
        <InfoSVG />
      </Icon>
      <Icon color="plum" disabled>
        <InfoSVG />
      </Icon>
      <Icon small>
        <InfoSVG />
      </Icon>
      <Icon smaller>
        <InfoSVG />
      </Icon>
    </>
  );
};
