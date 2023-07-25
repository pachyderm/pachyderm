import client from '../../client';

describe('services/version', () => {
  describe('getVersion', () => {
    it('should return version info', async () => {
      const pachClient = client();
      const version = await pachClient.version().getVersion();
      expect(version.major).toBeDefined();
    });
  });
});
