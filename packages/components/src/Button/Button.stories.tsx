import React from 'react';
import {BrowserRouter} from 'react-router-dom';

import {Button} from './';

export default {title: 'Button'};

export const Default = () => {
  return <Button autoSize>Primary Button</Button>;
};

export const Disabled = () => {
  return (
    <Button autoSize disabled={true}>
      Primary Button
    </Button>
  );
};

export const Secondary = () => {
  return (
    <Button autoSize buttonType="secondary">
      Secondary Button
    </Button>
  );
};

export const Tertiary = () => {
  return (
    <Button autoSize buttonType="tertiary">
      Tertiary Button
    </Button>
  );
};

export const Link = () => {
  return (
    <BrowserRouter>
      <Button autoSize to="/">
        Primary Link
      </Button>
    </BrowserRouter>
  );
};
