import {fixSchema} from '..';

describe('fixSchema', () => {
  it('moves a top-level referenced defintion to the top-level and deletes the ref', () => {
    const out = fixSchema({
      $ref: '#/definitions/foo',
      definitions: {
        foo: {
          find: 'me',
        },
      },
    });

    expect(out).toEqual({
      definitions: expect.any(Object),
      find: 'me',
    });
  });

  it('does not change an object that does not have a ref', () => {
    const out = fixSchema({
      foo: 'bar',
    });

    expect(out).toEqual({
      foo: 'bar',
    });
  });

  it.each([
    ['empty object', {}],
    ['null', null],
    ['undefined', undefined],
    ['number', 1],
    ['string', 'foobar'],
  ])('given a(n) %p it returns an empty object', (_, input) => {
    const out = fixSchema(input as any);

    expect(out).toEqual({});
  });
});
