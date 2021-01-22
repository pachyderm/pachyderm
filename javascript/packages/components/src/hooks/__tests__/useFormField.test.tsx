import {render, act} from '@testing-library/react';
import React from 'react';
import {useForm} from 'react-hook-form';

import useFormField from 'hooks/useFormField';
import {type, click} from 'testHelpers';

import {Button} from '../../Button';
import {Form} from '../../Form';
import {Label} from '../../Label';

describe('useFormField', () => {
  const FormField = () => {
    const {register, error: ErrorComponent, errorId, hasError} = useFormField(
      'test',
    );

    return (
      <>
        <Label htmlFor="test" label="Test" />
        <input
          aria-invalid={hasError}
          aria-describedby={errorId}
          name="test"
          id="test"
          ref={register({required: 'Test is required'})}
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
    const {findByLabelText, findByRole, queryByRole} = render(<TestBed />);

    const submitButton = await findByRole('button');
    const testInput = await findByLabelText('Test');

    await act(async () => {
      await click(submitButton);
    });

    expect(testInput).toHaveAttribute('aria-invalid', 'true');
    expect(testInput).toHaveAttribute('aria-describedby', 'test-error');
    expect(await findByRole('alert')).toBeInTheDocument();

    await type(testInput, 'a');
    await act(async () => {
      await click(submitButton);
    });

    expect(testInput).toHaveAttribute('aria-invalid', 'false');
    expect(testInput).not.toHaveAttribute('aria-describedby');
    expect(queryByRole('alert')).toBeNull();
  });
});
