import client from '../../client';

describe('services/admin', () => {
  describe('inspectCluster', () => {
    it('should return a cluster id', async () => {
      const pachClient = client({ssl: false, pachdAddress: 'localhost:30650'});
      const cluster = await pachClient.admin().inspectCluster();

      expect(cluster.id).toBeDefined();
    });
  });
});
