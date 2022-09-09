import noop from 'lodash/noop';
import React, {useMemo} from 'react';
import {
  FormProvider,
  SubmitHandler,
  UseFormReturn,
  FieldValues,
} from 'react-hook-form';

export interface FormProps<T extends FieldValues>
  extends Omit<React.FormHTMLAttributes<HTMLFormElement>, 'onSubmit'> {
  formContext: UseFormReturn<T>;
  onSubmit?: SubmitHandler<T>;
}

export const Form = <T extends FieldValues>({
  children,
  formContext,
  onSubmit = noop,
  ...rest
}: FormProps<T>) => {
  const handleSubmit = useMemo(
    () => formContext.handleSubmit(onSubmit),
    [formContext, onSubmit],
  );

  return (
    <FormProvider {...formContext}>
      <form onSubmit={handleSubmit} {...rest}>
        {children}
      </form>
    </FormProvider>
  );
};

export default Form;
