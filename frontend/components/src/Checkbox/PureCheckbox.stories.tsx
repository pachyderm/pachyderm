import {ComponentStory, ComponentMeta} from '@storybook/react';
import React, {useEffect, useState} from 'react';

import {PureCheckbox} from './Checkbox';

export default {
  title: 'Checkbox/Pure',
  component: PureCheckbox,
} as ComponentMeta<typeof PureCheckbox>;

const Template: ComponentStory<typeof PureCheckbox> = ({selected, ...args}) => {
  const [value, setValue] = useState(false);
  useEffect(() => setValue(selected), [selected]);
  const handleChange = () => setValue((value) => !value);

  return <PureCheckbox onChange={handleChange} selected={value} {...args} />;
};

const defaultArgs = {
  selected: false,
  disabled: false,
  small: false,
  id: 'checkbox',
  name: 'checkbox',
  label: 'I accept the terms and conditions',
};

export const RegularPureCheckbox = Template.bind({});
RegularPureCheckbox.args = {
  ...defaultArgs,
};

export const DisabledCheckbox = Template.bind({});
DisabledCheckbox.args = {
  ...defaultArgs,
  disabled: true,
};

export const SmallCheckbox = Template.bind({});
SmallCheckbox.args = {
  ...defaultArgs,
  small: true,
};
