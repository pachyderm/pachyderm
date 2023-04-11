import runJQFilter from '../../mock/utils/runJQFilter';
import {jqSelect, jqIn, jqCombine} from '../jqHelpers';

const trees = [
  {
    commonName: 'Oak',
    family: 'Fagaceae',
    order: 1,
  },
  {
    commonName: 'Maple',
    family: 'Sapindaceae',
    order: 2,
  },
  {
    commonName: 'Redwood',
    family: 'Cupressaceae',
    order: 2,
  },
  {
    commonName: 'Birch',
    family: 'Betulaceae',
    order: 1,
  },
  {
    commonName: 'Pine',
    family: 'Pinaceae',
    order: 1,
  },
];

describe('jqHelpers', () => {
  test('should construct a filter for items in a subset of strings', async () => {
    const filter = jqCombine(
      jqSelect(jqIn('.commonName', ['Oak', 'Maple', 'Birch'])),
    );
    expect(typeof filter).toBe('string');

    const result = await runJQFilter({
      jqFilter: `.trees[] | ${filter}`,
      object: {
        trees,
      },
      objectMapper: (i) => i,
    });

    expect(result).toHaveLength(3);
    expect(result[0].commonName).toBe('Oak');
    expect(result[1].commonName).toBe('Maple');
    expect(result[2].commonName).toBe('Birch');
  });

  test('should construct a filter for items in a subset of numbers', async () => {
    const filter = jqCombine(jqSelect(jqIn('.order', [2])));
    const result = await runJQFilter({
      jqFilter: `.trees[] | ${filter}`,
      object: {
        trees,
      },
      objectMapper: (i) => i,
    });
    expect(typeof filter).toBe('string');

    expect(result).toHaveLength(2);
    expect(result[0].order).toBe(2);
    expect(result[1].order).toBe(2);
  });

  test('should construct a filter with multiple subsets', async () => {
    let filter = '';
    filter = jqCombine(
      filter,
      jqSelect(jqIn('.commonName', ['Oak', 'Maple', 'Birch'])),
    );
    filter = jqCombine(
      filter,
      jqSelect(jqIn('.family', ['Sapindaceae', 'Betulaceae'])),
    );
    expect(typeof filter).toBe('string');

    const result = await runJQFilter({
      jqFilter: `.trees[] | ${filter}`,
      object: {
        trees,
      },
      objectMapper: (i) => i,
    });

    expect(result).toHaveLength(2);
    expect(result[0].commonName).toBe('Maple');
    expect(result[1].commonName).toBe('Birch');
    expect(result[0].family).toBe('Sapindaceae');
    expect(result[1].family).toBe('Betulaceae');
  });
});
