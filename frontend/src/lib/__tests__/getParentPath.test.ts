import getParentPath from '@dash-frontend/lib/getParentPath';

describe('lib/getParentPath', () => {
  it('should get the parent path', () => {
    expect(getParentPath()).toBe('/');
    expect(getParentPath('')).toBe('/');
    expect(getParentPath('/')).toBe('/');
    expect(getParentPath('/hello')).toBe('/');
    expect(getParentPath('/hello/')).toBe('/');
    expect(getParentPath('/hello/world')).toBe('/hello/');
    expect(getParentPath('/hello/world/')).toBe('/hello/');
    expect(getParentPath('/hello/world/again/')).toBe('/hello/world/');
    expect(getParentPath('/hello/world/again')).toBe('/hello/world/');
  });
});
