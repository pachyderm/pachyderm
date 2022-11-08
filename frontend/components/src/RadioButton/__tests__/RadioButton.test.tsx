import {render} from '@testing-library/react';
import React from 'react';
import {useForm, RegisterOptions} from 'react-hook-form';

import {click} from '@dash-frontend/testHelpers';
import {Button, Form} from '@pachyderm/components';

import RadioButton from '../RadioButton';

describe('Radio Button', () => {
  const onSubmit = jest.fn();

  const TestBed = ({
    validationOptions = {},
  }: {
    validationOptions?: RegisterOptions;
  }) => {
    const formCtx = useForm();

    return (
      <Form formContext={formCtx} onSubmit={onSubmit}>
        <RadioButton
          id="yes"
          value="yes"
          name="answer"
          validationOptions={validationOptions}
        >
          <RadioButton.Label>Yes</RadioButton.Label>
        </RadioButton>

        <RadioButton
          id="no"
          value="no"
          name="answer"
          validationOptions={validationOptions}
        >
          <RadioButton.Label>No</RadioButton.Label>
        </RadioButton>

        <Button type="submit">Submit</Button>
      </Form>
    );
  };

  beforeEach(() => {
    jest.resetAllMocks();
  });

  it('should allow user to select an option from the radio group', async () => {
    const {getByLabelText, getByRole} = render(<TestBed />);

    const yes = getByLabelText('Yes');
    const no = getByLabelText('No');
    const submit = getByRole('button');

    await click(yes);
    await click(submit);

    expect(onSubmit.mock.calls[0][0]).toStrictEqual({answer: 'yes'});

    await click(no);
    await click(submit);

    expect(onSubmit.mock.calls[1][0]).toStrictEqual({answer: 'no'});
  });

  it('should accept validation options', async () => {
    const {getByLabelText, getByRole} = render(
      <TestBed validationOptions={{required: true}} />,
    );

    const yes = getByLabelText('Yes');
    const submit = getByRole('button');

    await click(submit);

    expect(onSubmit).not.toHaveBeenCalled();

    await click(yes);
    await click(submit);

    expect(onSubmit).toHaveBeenCalled();
  });
});
