import React from 'react';

import ShortId from '@dash-frontend/components/ShortId';
import {SHORT_OR_LONG_UUID} from '@dash-frontend/constants/pachCore';

const extractAndShortenIds = (inputString: string) => {
  return (
    <span>
      {inputString.split(' ').map((substring) =>
        substring.match(SHORT_OR_LONG_UUID) ? (
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
