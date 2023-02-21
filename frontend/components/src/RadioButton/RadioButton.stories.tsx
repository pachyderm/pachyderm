import React, {useState} from 'react';
import {useForm} from 'react-hook-form';

import {Form} from '../Form';

import {RadioButton} from './';

export default {title: 'RadioButton'};

export const Default = () => {
  const formCtx = useForm();

  return (
    <Form formContext={formCtx}>
      <RadioButton id="yes" name="answer" value="yes">
        <RadioButton.Label>Yes</RadioButton.Label>
      </RadioButton>
      <RadioButton id="no" name="answer" value="no">
        <RadioButton.Label>No</RadioButton.Label>
      </RadioButton>
      <RadioButton id="disabled" name="answer" value="disabled" disabled>
        <RadioButton.Label>Disabled</RadioButton.Label>
      </RadioButton>
      <RadioButton
        id="small-disabled"
        name="answer"
        value="small-disabled"
        disabled
        small
      >
        <RadioButton.Label>Small Disabled</RadioButton.Label>
      </RadioButton>
    </Form>
  );
};

export const Pure = () => {
  const [value, setValue] = useState('');

  return (
    <>
      <RadioButton.Pure
        id="yes"
        name="answer"
        value="yes"
        selected={value === 'yes'}
        onChange={() => setValue('yes')}
      >
        <RadioButton.Label>Yes</RadioButton.Label>
      </RadioButton.Pure>
      <RadioButton.Pure
        id="no"
        name="answer"
        value="no"
        selected={value === 'no'}
        onChange={() => setValue('no')}
      >
        <RadioButton.Label>No</RadioButton.Label>
      </RadioButton.Pure>
      <RadioButton.Pure
        id="disabled"
        name="answer"
        value="disabled"
        disabled
        selected={value === 'disabled'}
        onChange={() => setValue('disabled')}
      >
        <RadioButton.Label>Disabled</RadioButton.Label>
      </RadioButton.Pure>
      <RadioButton.Pure
        id="small-disabled"
        name="answer"
        value="small-disabled"
        disabled
        selected={value === 'small-disabled'}
        onChange={() => setValue('small-disabled')}
        small
      >
        <RadioButton.Label>Small Disabled</RadioButton.Label>
      </RadioButton.Pure>
    </>
  );
};
