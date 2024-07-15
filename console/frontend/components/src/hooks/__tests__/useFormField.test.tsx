import {render, act, screen} from '@testing-library/react';
import React from 'react';
import {useForm} from 'react-hook-form';

import {type, click} from 'testHelpers';

import {Button} from '../../Button';
import {Form} from '../../Form';
import {Label} from '../../Label';
import useFormField from '../useFormField';

describe('useFormField', () => {
  const FormField = () => {
    const {
      register,
      error: ErrorComponent,
      errorId,
      hasError,
    } = useFormField('test');

    return (
      <>
        <Label htmlFor="test" label="Test" />
        <input
          aria-invalid={hasError}
          aria-describedby={errorId}
          id="test"
          {...register('test', {required: 'Test is required'})}
        />
        <ErrorComponent />
      </>
    );
  };

  const TestBed = () => {
    const formCtx = useForm();

    return (
      <Form formContext={formCtx}>
        <FormField />

        <Button type="submit">Submit</Button>
      </Form>
    );
  };

  it('should correctly generate error messages, and apply aria attributes', async () => {
    render(<TestBed />);

    const submitButton = await screen.findByRole('button');
    const testInput = await screen.findByLabelText('Test');

    await act(async () => {
      await click(submitButton);
    });

    expect(testInput).toHaveAttribute('aria-invalid', 'true');
    expect(testInput).toHaveAttribute('aria-describedby', 'test-error');
    expect(await screen.findByRole('alert')).toBeInTheDocument();

    await type(testInput, 'a');
    await act(async () => {
      await click(submitButton);
    });

    expect(testInput).toHaveAttribute('aria-invalid', 'false');
    expect(testInput).not.toHaveAttribute('aria-describedby');
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });
});
