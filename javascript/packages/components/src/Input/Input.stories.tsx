import React, {useEffect} from 'react';
import {useForm} from 'react-hook-form';

import {Form} from '../Form';
import {Label} from '../Label';

import {Input} from './';

export default {title: 'Input'};

interface FormValues {
  name: string;
  email: string;
}

export const Default = () => {
  const formCtx = useForm<FormValues>({mode: 'onChange'});

  return (
    <Form formContext={formCtx}>
      <Input name="name" id="name" />
    </Form>
  );
};

export const Disabled = () => {
  const formCtx = useForm<FormValues>({mode: 'onChange'});

  return (
    <Form formContext={formCtx}>
      <Input name="disabled" disabled value="Disabled Input" />
    </Form>
  );
};

export const ValidationError = () => {
  const formCtx = useForm<FormValues>();
  const {setError} = formCtx;

  useEffect(() => {
    setError('name', {type: 'error', message: 'This is an error message.'});
  }, [setError]);
  return (
    <Form formContext={formCtx}>
      <Input name="name" readOnly value="Disabled Input" />
    </Form>
  );
};

export const WithLabel = () => {
  const formCtx = useForm<FormValues>();

  return (
    <Form formContext={formCtx}>
      <Label htmlFor="name" label="Name" />
      <Input name="name" placeholder="Peter" />
    </Form>
  );
};
