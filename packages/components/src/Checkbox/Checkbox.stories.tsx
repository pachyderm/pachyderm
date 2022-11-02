import {ComponentStory, ComponentMeta} from '@storybook/react';
import React, {useEffect} from 'react';

import {PureCheckbox} from './Checkbox';

export default {
  title: 'Checkbox',
  component: PureCheckbox,
} as ComponentMeta<typeof PureCheckbox>;

const Template: ComponentStory<typeof PureCheckbox> = ({selected, ...args}) => {
  const [value, setValue] = React.useState(false);
  useEffect(() => setValue(selected), [selected]);
  const handleChange = () => setValue((value) => !value);

  return <PureCheckbox onChange={handleChange} selected={value} {...args} />;
};

const defaultArgs = {
  selected: true,
  disabled: false,
  small: false,
  id: 'checkbox',
  name: 'checkbox',
  label: 'I accept the terms and conditions',
};

export const Checkbox = Template.bind({});
Checkbox.args = {
  ...defaultArgs,
};

export const Disabled = Template.bind({});
Disabled.args = {
  ...defaultArgs,
  disabled: true,
};

export const Small = Template.bind({});
Small.args = {
  ...defaultArgs,
  small: true,
};
