import React from 'react';
import {useForm} from 'react-hook-form';

import {Form} from '../Form';
import {Input} from '../Input';

import {Label} from './';

export default {title: 'Label'};

interface FormValues {
  name: string;
  email: string;
}

export const Default = () => {
  return (
    <>
      <Label htmlFor="name" label="Name" />
    </>
  );
};

export const MaxLength = () => {
  const formCtx = useForm<FormValues>();
  return (
    <Form formContext={formCtx}>
      <Label htmlFor="name" label="Name" maxLength={20} />
      <Input
        name="name"
        validationOptions={{
          maxLength: {
            value: 20,
            message: 'Name is too long',
          },
        }}
      />
    </Form>
  );
};
