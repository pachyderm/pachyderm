import client from '../../client';

describe('services/admin', () => {
  describe('inspectCluster', () => {
    it('should return a cluster id', async () => {
      const pachClient = client();
      const cluster = await pachClient.admin().inspectCluster();

      expect(cluster.id).toBeDefined();
    });
  });
});
