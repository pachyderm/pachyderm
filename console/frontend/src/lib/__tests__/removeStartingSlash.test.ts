import removeStartingSlash from '@dash-frontend/lib/removeStartingSlash';

describe('lib/removeStartingSlash', () => {
  it('should remove starting slash', () => {
    expect(removeStartingSlash()).toBe('/');
    expect(removeStartingSlash('')).toBe('/');
    expect(removeStartingSlash('/')).toBe('/');
    expect(removeStartingSlash('/hello')).toBe('hello');
    expect(removeStartingSlash('/hello/')).toBe('hello/');
    expect(removeStartingSlash('/hello/world')).toBe('hello/world');
    expect(removeStartingSlash('///hello/world/')).toBe('//hello/world/');
    expect(removeStartingSlash('.hello/world/')).toBe('.hello/world/');
  });
});
