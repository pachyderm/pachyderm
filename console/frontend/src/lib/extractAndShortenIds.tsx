import React from 'react';

import GlobalIdCopy from '@dash-frontend/components/GlobalIdCopy';
import {SHORT_OR_LONG_UUID} from '@dash-frontend/constants/pachCore';

const extractAndShortenIds = (inputString: string) => {
  return (
    <span>
      {inputString.split(' ').map((substring) =>
        substring.match(SHORT_OR_LONG_UUID) ? (
          <span key={substring}>
            <GlobalIdCopy shortenId id={substring} />{' '}
          </span>
        ) : (
          `${substring} `
        ),
      )}
    </span>
  );
};

export default extractAndShortenIds;
