import {render, waitFor, act} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import {useForm} from 'react-hook-form';

import {Button} from 'Button';
import {Form} from 'Form';

import {Checkbox, CheckboxProps} from '../';

describe('Checkbox', () => {
  jest.mock('Svg', () => {
    return {};
  });

  const onSubmit = jest.fn();

  const TestBed = (props: CheckboxProps) => {
    const formCtx = useForm();

    return (
      <Form formContext={formCtx} onSubmit={onSubmit}>
        <Checkbox {...props} />

        <Button type="submit">Submit</Button>
      </Form>
    );
  };

  beforeEach(() => {
    jest.resetAllMocks();
  });

  it('should be interactive', async () => {
    const {getByLabelText, getByText} = render(
      <TestBed id="test" name="test" label="Test Checkbox Label" />,
    );

    const checkbox = getByLabelText('Test Checkbox Label');
    const submit = getByText('Submit');

    await act(async () => {
      await userEvent.click(checkbox);
    });

    await act(async () => {
      await userEvent.click(submit);
    });

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));

    expect(onSubmit.mock.calls[0][0]).toStrictEqual({test: true});

    await act(async () => {
      await userEvent.click(checkbox);
    });

    await act(async () => {
      await userEvent.click(submit);
    });

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(2));
    expect(onSubmit.mock.calls[1][0]).toStrictEqual({test: false});
  });

  it('should accept validation options', async () => {
    const {getByLabelText, getByText} = render(
      <TestBed
        id="test"
        name="test"
        label="Test Checkbox Label"
        validationOptions={{required: true}}
      />,
    );

    const checkbox = getByLabelText('Test Checkbox Label');
    const submit = getByText('Submit');

    await act(async () => {
      await userEvent.click(submit);
    });

    await act(async () => {
      await userEvent.click(checkbox);
    });

    await act(async () => {
      await userEvent.click(submit);
    });

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit.mock.calls[0][0]).toStrictEqual({test: true});
  });
});
