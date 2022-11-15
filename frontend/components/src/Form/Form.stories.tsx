import React, {useState} from 'react';
import {useForm} from 'react-hook-form';

import {Button} from '../Button';
import {Input} from '../Input';
import {Label} from '../Label';

import {Form} from './';

export default {title: 'Form'};

interface FormValues {
  name: string;
  email: string;
}

export const Default = () => {
  const [submittedValues, setSubmittedValues] = useState<FormValues>();

  const handleSubmit = (values: FormValues) => {
    setSubmittedValues(values);
  };

  const formCtx = useForm<FormValues>();
  return (
    <Form formContext={formCtx} onSubmit={handleSubmit}>
      <div style={{marginBottom: '1rem'}}>
        <Label htmlFor="name" label="Name" />
        <Input name="name" validationOptions={{required: 'Required'}} />
      </div>
      <div style={{marginBottom: '1rem'}}>
        <Label htmlFor="email" label="Email" />
        <Input name="email" validationOptions={{required: 'Required'}} />
      </div>
      <Button type="submit">Submit</Button>
      <h3>Submitted Values</h3>
      <p>Name: {submittedValues?.name}</p>
      <p>Email: {submittedValues?.email}</p>
    </Form>
  );
};
