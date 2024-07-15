import React from 'react';
import {useForm} from 'react-hook-form';

import {Form} from '../Form';

import {TagsInput} from './';

export default {title: 'TagsInput'};

const TagsInputWrapper = ({...props}) => {
  const formContext = useForm();

  return (
    <Form formContext={formContext}>
      <TagsInput style={{width: '400px'}} {...props} />
    </Form>
  );
};

export const Default = () => {
  return <TagsInputWrapper />;
};

export const WithValidation = () => {
  return (
    <TagsInputWrapper
      defaultValues={[
        'cam@pachyderm.io',
        'invalid',
        'connor@pachyderm.io',
        'jeff@pachyderm.io',
      ]}
      validate={(v: string) => v.endsWith('@pachyderm.io')}
    />
  );
};
