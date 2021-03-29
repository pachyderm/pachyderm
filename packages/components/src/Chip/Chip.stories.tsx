import React from 'react';
import {useForm} from 'react-hook-form';

import {Form} from '../Form';

import {ChipGroup} from './Chip';

import {Chip, ChipInput} from './';

export default {title: 'Chip'};

export const Default = () => {
  return <Chip>Jobs (3)</Chip>;
};

export const Input = () => {
  const formCtx = useForm({mode: 'onChange'});
  return (
    <Form formContext={formCtx}>
      <ChipInput name="chipInput" label="Chip Input" />
    </Form>
  );
};

export const Group = () => {
  const formCtx = useForm({mode: 'onChange'});
  return (
    <Form formContext={formCtx}>
      <div style={{width: '25rem'}}>
        <ChipGroup>
          <ChipInput name="Failure" />
          <ChipInput name="Starting" />
          <ChipInput name="Thinking" />
          <ChipInput name="Running" />
          <ChipInput name="Success" />
          <ChipInput name="Killed" />
          <ChipInput name="Egressing" />
        </ChipGroup>
      </div>
    </Form>
  );
};
