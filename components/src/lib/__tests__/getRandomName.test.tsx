import {PACHYDERMS, ADJECTIVES} from '../constants/names';
import getRandomName from '../getRandomName';

describe('getRandomName', () => {
  it('should generate a cute pachyderm name', () => {
    const name = getRandomName();
    const [adjective, pachyderm] = name.split('-');

    expect(PACHYDERMS).toContain(pachyderm);
    expect(ADJECTIVES).toContain(adjective);
  });
});
