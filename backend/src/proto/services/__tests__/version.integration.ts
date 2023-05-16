import client from '../../client';

describe('services/version', () => {
  describe('getVersion', () => {
    it('should return version info', async () => {
      const pachClient = client({ssl: false, pachdAddress: 'localhost:30650'});
      const version = await pachClient.version().getVersion();
      expect(version.major).toBeDefined();
    });
  });
});
