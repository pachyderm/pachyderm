import {verifyAuthorizationStatus} from '../useVerifiedAuthorization';

describe('verifyAuthorizationStatus', () => {
  const cases: [boolean | null | undefined, boolean][] = [
    [true, true],
    [false, false],
    [null, true],
    [undefined, true],
  ];
  test.each(cases)('%p returns %p', (input, result) => {
    expect(verifyAuthorizationStatus(input)).toBe(result);
  });
});
