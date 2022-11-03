import React from 'react';
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
