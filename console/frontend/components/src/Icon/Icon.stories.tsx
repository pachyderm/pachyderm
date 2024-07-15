import React from 'react';

import {InfoSVG} from '../Svg';

import {Icon} from './';

export default {title: 'Icon'};

export const Default = () => {
  return (
    <>
      <Icon color="black" hoverColor="red">
        <InfoSVG />
      </Icon>
      <Icon color="white" hoverColor="black">
        <InfoSVG />
      </Icon>
      <Icon color="plum" hoverColor="grey">
        <InfoSVG />
      </Icon>
      <Icon color="grey" hoverColor="black">
        <InfoSVG />
      </Icon>
      <Icon color="green" hoverColor="black">
        <InfoSVG />
      </Icon>
      <Icon color="blue" hoverColor="black">
        <InfoSVG />
      </Icon>
      <Icon color="red" hoverColor="black">
        <InfoSVG />
      </Icon>
      <Icon color="yellow" hoverColor="black">
        <InfoSVG />
      </Icon>
      <Icon color="highlightGreen" hoverColor="black">
        <InfoSVG />
      </Icon>
      <Icon color="highlightOrange" hoverColor="black">
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
