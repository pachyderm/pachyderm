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
    </Form>
  );
};
