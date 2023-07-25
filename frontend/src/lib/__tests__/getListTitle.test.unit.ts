import getListTitle from '@dash-frontend/lib/getListTitle';

describe('getListTitle', () => {
  const cases: [string, number, string][] = [
    ['Job', 0, 'No Jobs Found'],
    ['Job', 1, 'Last Job'],
    ['Job', 2, 'Last 2 Jobs'],
  ];
  test.each(cases)('%p returns %p', (input1, input2, result) => {
    expect(getListTitle(input1, input2)).toBe(result);
  });
});
