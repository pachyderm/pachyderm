import React from 'react';
import {useForm} from 'react-hook-form';

import {Form} from '../Form';

import {Select} from './';

export default {title: 'Select'};

export const WithPlaceholder = () => {
  const form = useForm();

  return (
    <Form formContext={form}>
      <Select id="select">
        <Select.Option value="1">Option 1</Select.Option>
        <Select.Option value="2">Option 2</Select.Option>
        <Select.Option value="3">Option 3</Select.Option>
      </Select>
    </Form>
  );
};

export const WithInitialValue = () => {
  const form = useForm();

  return (
    <Form formContext={form}>
      <Select id="select" initialValue="1">
        <Select.Option value="1">Option 1</Select.Option>
        <Select.Option value="2">Option 2</Select.Option>
        <Select.Option value="3">Option 3</Select.Option>
      </Select>
    </Form>
  );
};
