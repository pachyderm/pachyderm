import React from 'react';

import ShortId from '@dash-frontend/components/ShortId';

const extractAndShortenIds = (inputString: string) => {
  return (
    <>
      {inputString
        .split(' ')
        .map((substring) =>
          substring.match(/[a-zA-Z0-9_-]{64}|[a-zA-Z0-9_-]{32}/) ? (
            <ShortId inputString={substring} />
          ) : (
            `${substring} `
          ),
        )}
    </>
  );
};

export default extractAndShortenIds;
