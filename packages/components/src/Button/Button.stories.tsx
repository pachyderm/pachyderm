import React from 'react';
import {BrowserRouter} from 'react-router-dom';

import {ButtonGroup, ButtonProps} from 'index';
import {UploadSVG} from 'Svg';

import {Button} from './';

export default {title: 'Button', component: Button};

const defaultArgTypes = {
  href: {control: {disable: true}},
  to: {control: {disable: true}},
  'data-testid': {control: {disable: true}},
  download: {control: {disable: true}},
  disabled: {control: {type: 'boolean'}},
  color: {control: {type: 'text'}},
  buttonType: {
    control: {
      type: 'radio',
      options: ['primary', 'secondary', 'ghost', 'tertiary', 'dropdown'],
    },
  },
  iconPosition: {
    control: {
      type: 'radio',
      options: ['both', 'start', 'end'],
    },
  },
};

export const Default = () => {
  return <Button>Button Text</Button>;
};

export const TextButton = (args: ButtonProps) => {
  return <Button {...args}>Button Text</Button>;
};

TextButton.argTypes = {
  IconSVG: {control: {disable: true}},
  ...defaultArgTypes,
};

export const IconButton = (args: ButtonProps) => {
  return <Button {...args} IconSVG={UploadSVG} />;
};

IconButton.argTypes = {
  IconSVG: {control: {disable: true}},
  ...defaultArgTypes,
};

export const TextIconButton = (args: ButtonProps) => {
  return (
    <Button {...args} IconSVG={UploadSVG}>
      Button Text
    </Button>
  );
};

TextIconButton.argTypes = {
  icon: {control: {disable: true}},
  ...defaultArgTypes,
};

export const Link = () => {
  return (
    <BrowserRouter>
      <Button to="/">Primary Link</Button>
    </BrowserRouter>
  );
};

export const MultiButtonGroup = () => {
  return (
    <ButtonGroup>
      <Button buttonType="ghost">Tertiary Button</Button>
      <Button buttonType="ghost">Tertiary Button</Button>
      <Button>Button Text</Button>
      <Button buttonType="secondary" IconSVG={UploadSVG}>
        Button Text
      </Button>
      <Button IconSVG={UploadSVG} />
    </ButtonGroup>
  );
};
