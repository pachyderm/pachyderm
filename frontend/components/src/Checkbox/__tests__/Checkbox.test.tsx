import {render, waitFor, act} from '@testing-library/react';
import React from 'react';
import {useForm} from 'react-hook-form';

import {click} from '@dash-frontend/testHelpers';
import {Button, Form} from '@pachyderm/components';

import {Checkbox, CheckboxProps} from '../';

type TestBedProps = CheckboxProps & {
  onSubmit: () => void;
};

describe('Checkbox', () => {
  jest.mock('Svg', () => {
    return {};
  });

  const TestBed = (props: TestBedProps) => {
    const formCtx = useForm();

    return (
      <Form formContext={formCtx} onSubmit={props.onSubmit}>
        <Checkbox {...props} />

        <Button type="submit">Submit</Button>
      </Form>
    );
  };

  it('should be interactive', async () => {
    const onSubmit = jest.fn();

    const {getByLabelText, getByText} = render(
      <TestBed
        id="test"
        name="test"
        label="Test Checkbox Label"
        onSubmit={onSubmit}
      />,
    );

    const checkbox = getByLabelText('Test Checkbox Label');
    const submit = getByText('Submit');

    await click(checkbox);
    await click(submit);

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));

    expect(onSubmit.mock.calls[0][0]).toStrictEqual({test: true});

    await act(async () => {
      await click(checkbox);
    });

    await click(submit);

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(2));
    expect(onSubmit.mock.calls[1][0]).toStrictEqual({test: false});
  });

  it('should accept validation options', async () => {
    const onSubmit = jest.fn();

    const {getByLabelText, getByText} = render(
      <TestBed
        id="test"
        name="test"
        label="Test Checkbox Label"
        validationOptions={{required: true}}
        onSubmit={onSubmit}
      />,
    );

    const checkbox = getByLabelText('Test Checkbox Label');
    const submit = getByText('Submit');

    await click(submit);
    await click(checkbox);
    await click(submit);

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit.mock.calls[0][0]).toStrictEqual({test: true});
  });

  it('should be disabled', async () => {
    const onSubmit = jest.fn();

    const {getByLabelText, getByText} = render(
      <TestBed
        id="test"
        name="test"
        label="Test Checkbox Label"
        disabled={true}
        onSubmit={onSubmit}
      />,
    );

    const checkbox = getByLabelText('Test Checkbox Label');
    const submit = getByText('Submit');

    await click(checkbox);
    await click(submit);

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit.mock.calls[0][0]).toStrictEqual({test: undefined});
  });
});
