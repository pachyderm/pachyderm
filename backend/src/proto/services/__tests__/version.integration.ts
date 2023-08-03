import apiClientRequestWrapper from '../../client';

describe('services/version', () => {
  describe('getVersion', () => {
    it('should return version info', async () => {
      const pachClient = apiClientRequestWrapper();
      const version = await pachClient.version.getVersion();
      expect(version.major).toBeDefined();
    });
  });
});
