import isEmpty from 'lodash/isEmpty';

import omitByDeep from '../omitByDeep';

describe('omitDeep', () => {
  it('should omit deep values based on omitFn', () => {
    const deepJson = {
      key1: '',
      key2: undefined,
      key3: {
        key4: 'stuff',
        key5: undefined,
        key6: [
          {
            key7: undefined,
            key8: [],
            key9: [
              {
                key10: null,
                key11: 'more stuff',
              },
            ],
            key10: {},
          },
        ],
      },
    };

    expect(
      omitByDeep(
        deepJson,
        (val, _) => !val || (typeof val === 'object' && isEmpty(val)),
      ),
    ).toEqual({
      key3: {
        key4: 'stuff',
        key6: [
          {
            key9: [
              {
                key11: 'more stuff',
              },
            ],
          },
        ],
      },
    });
  });
});
