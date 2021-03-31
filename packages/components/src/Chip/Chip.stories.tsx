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
      <ChipInput name="chipInput">Chip Input</ChipInput>
    </Form>
  );
};

export const Group = () => {
  const formCtx = useForm({mode: 'onChange'});
  return (
    <Form formContext={formCtx}>
      <div style={{width: '25rem'}}>
        <ChipGroup>
          <ChipInput name="failure">Failure</ChipInput>
          <ChipInput name="starting">Starting</ChipInput>
          <ChipInput name="thinking">Thinking</ChipInput>
          <ChipInput name="running">Running</ChipInput>
          <ChipInput name="success">Success</ChipInput>
          <ChipInput name="killed">Killed</ChipInput>
          <ChipInput name="egressing">Egressing</ChipInput>
        </ChipGroup>
      </div>
    </Form>
  );
};
