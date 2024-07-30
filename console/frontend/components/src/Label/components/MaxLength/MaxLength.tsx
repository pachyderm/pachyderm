import React from 'react';
import {FieldPath, FieldValues, useFormContext} from 'react-hook-form';

import HelperText from '../HelperText';

interface MaxLengthProps {
  htmlFor: FieldPath<FieldValues>;
  maxLength: number;
  className?: string;
}

const MaxLength: React.FC<MaxLengthProps> = ({
  htmlFor,
  maxLength,
  className,
}) => {
  const {watch} = useFormContext();

  const value = watch(htmlFor);

  const remainingCharacters = maxLength - (value ? value.length : 0);

  return <HelperText className={className}>{remainingCharacters}</HelperText>;
};

export default MaxLength;
