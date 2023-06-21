import {AllRoles, safeHasAtLeastRole} from '../rbac';

describe('safeHasAtLeastRole', () => {
  const cases: [AllRoles[], AllRoles, boolean][] = [
    // Currently there would be around 800 test cases
    [['repoReader'], 'clusterAdmin', false],
    [['clusterAdmin'], 'clusterAdmin', true],
    [['projectOwner', 'repoReader'], 'projectViewer', true],
    [['projectViewer', 'repoReader'], 'projectViewer', true],
    [['projectWriter', 'repoReader'], 'projectViewer', true],
    [['projectOwner', 'repoReader'], 'projectOwner', true],
    [['projectViewer', 'repoReader'], 'projectOwner', false],
    [['projectOwner', 'repoReader'], 'repoWriter', false],
    [['projectViewer', 'repoWriter'], 'repoReader', true],
    [['projectWriter', 'repoOwner'], 'repoWriter', true],
    [['projectOwner', 'repoOwner'], 'projectOwner', true],
  ];

  test.each(cases)('%p at least has %p, returns %p', (first, second, third) => {
    const result = safeHasAtLeastRole(second, first);
    expect(result).toEqual(third);
  });

  test('it handles empty values', () => {
    expect(safeHasAtLeastRole('repoReader', undefined)).toBe(true); // This means auth is disabled
    expect(safeHasAtLeastRole('repoReader', [])).toBe(false);
  });
});
