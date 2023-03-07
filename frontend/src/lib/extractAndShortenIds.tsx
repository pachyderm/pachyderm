import React from 'react';

import ShortId from '@dash-frontend/components/ShortId';

const extractAndShortenIds = (inputString: string) => {
  return (
    <span>
      {inputString.split(' ').map((substring) =>
        substring.match(/[a-zA-Z0-9_-]{64}|[a-zA-Z0-9_-]{32}/) ? (
          <span key={substring}>
            <ShortId key={substring} inputString={substring} />{' '}
          </span>
        ) : (
          `${substring} `
        ),
      )}
    </span>
  );
};

export default extractAndShortenIds;
