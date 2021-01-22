import React, {useEffect} from 'react';
import {useForm} from 'react-hook-form';

import {Form} from '../Form';
import {Label} from '../Label';

import TextArea from './TextArea';

export default {title: 'TextArea'};

interface FormValues {
  name: string;
  email: string;
}

const placeholder =
  'Fusce tortor ipsum, varius quis arcu at, molestie pellentesque lacus. Aliquam metus magna, sagittis vitae quam et, scelerisque bibendum leo. Praesent tempor dignissim augue, a eleifend nibh congue eu. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Cras in nibh ac justo volutpat interdum. Duis finibus ligula tempor felis dictum efficitur. Nam efficitur sodales mi et gravida. Phasellus ut odio volutpat, interdum massa id, rutrum dui.';

export const Default = () => {
  const formCtx = useForm<FormValues>({mode: 'onChange'});

  return (
    <Form formContext={formCtx}>
      <TextArea style={{height: '7rem'}} name="name" id="name" />
    </Form>
  );
};

export const Disabled = () => {
  const formCtx = useForm<FormValues>({mode: 'onChange'});

  return (
    <Form formContext={formCtx}>
      <TextArea
        style={{height: '7rem'}}
        name="disabled"
        disabled
        value={placeholder}
      />
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
      <TextArea
        style={{height: '7rem'}}
        name="name"
        readOnly
        value="Disabled Input"
      />
    </Form>
  );
};

export const WithLabel = () => {
  const formCtx = useForm<FormValues>();

  return (
    <Form formContext={formCtx}>
      <Label htmlFor="name" label="Name" />
      <TextArea
        style={{height: '7rem'}}
        name="name"
        placeholder={placeholder}
      />
    </Form>
  );
};
