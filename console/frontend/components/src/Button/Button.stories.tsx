import React from 'react';
import {BrowserRouter} from 'react-router-dom';

import {UploadSVG, ButtonGroup, ButtonProps} from '@pachyderm/components';

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
      options: [
        'primary',
        'secondary',
        'ghost',
        'tertiary',
        'quaternary',
        'dropdown',
      ],
    },
  },
  iconPosition: {
    control: {
      type: 'radio',
      options: ['both', 'start', 'end'],
    },
  },
  IconSVG: {control: {disable: true}},
};

export const Default = (args: ButtonProps) => {
  return <Button {...args}>Button Text</Button>;
};

Default.argTypes = defaultArgTypes;

export const IconButton = (args: ButtonProps) => {
  return <Button {...args} IconSVG={UploadSVG} />;
};

IconButton.argTypes = defaultArgTypes;

export const TextIconButton = (args: ButtonProps) => {
  return (
    <Button {...args} IconSVG={UploadSVG}>
      Button Text
    </Button>
  );
};

TextIconButton.argTypes = defaultArgTypes;

export const Link = (args: ButtonProps) => {
  return (
    <BrowserRouter>
      <Button to="/" {...args}>
        Primary Link
      </Button>
    </BrowserRouter>
  );
};

Link.argTypes = defaultArgTypes;

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
