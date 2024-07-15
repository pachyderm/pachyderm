import {UUID_WITHOUT_DASHES_REGEX} from '@dash-frontend/constants/pachCore';
import generateId from '@dash-frontend/lib/generateId';

describe('generateId', () => {
  it('should match UUID regex', () => {
    const id = generateId();
    expect(id).toMatch(UUID_WITHOUT_DASHES_REGEX);
  });
});
