import React from 'react';
import {useForm} from 'react-hook-form';

import {Form} from '../Form';

import {PureCheckbox} from './Checkbox';

import {Checkbox} from './';

export default {title: 'Checkbox'};

interface FormValues {
  checkbox: string;
}

export const Default = () => {
  const formCtx = useForm<FormValues>({mode: 'onChange'});

  return (
    <Form formContext={formCtx}>
      <Checkbox
        id="checkbox"
        name="checkbox"
        label="I accept the terms and conditions"
      />
    </Form>
  );
};

export const disabledCheckbox = () => {
  return (
    <>
      <PureCheckbox
        selected={false}
        disabled
        id="checkbox"
        name="checkbox"
        label="I accept the terms and conditions"
      />
    </>
  );
};
