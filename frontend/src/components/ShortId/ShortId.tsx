import React from 'react';

import CopiableField from './CopiableField';

interface ShortIdProps {
  inputString: string;
  error?: boolean;
}

const ShortId: React.FC<ShortIdProps> = ({inputString, error}) => {
  const shortId = inputString.slice(0, 8);
  inputString.replace(/[a-zA-Z0-9_-]{64}|[a-zA-Z0-9_-]{32}/, 'ID');

  return (
    <CopiableField
      inputString={shortId}
      inline={true}
      successCheckmarkAriaLabel="You have successfully copied the id"
    >
      {shortId}
    </CopiableField>
  );
};

export default ShortId;
