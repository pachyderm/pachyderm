import client from '../../client';
import {State} from '../../proto/enterprise/enterprise_pb';

describe('enterprise', () => {
  afterEach(async () => {
    const pachClient = client({ssl: false, pachdAddress: 'localhost:30650'});
    await pachClient.license().deleteAll();
  });
  it('should activate enterprise', async () => {
    const enterpriseKey = process.env.ENTERPRISE_KEY || '';
    const pachClient = client({ssl: false, pachdAddress: 'localhost:30650'});
    const enterprise = pachClient.enterprise();
    const license = pachClient.license();
    await license.activateLicense(enterpriseKey);
    await license.addCluster('localhost', 'localhost:1650', 'secret');
    await license.updateCluster(
      'localhost',
      'localhost:1650',
      'localhost:16650',
    );
    await enterprise.activate('localhost:1650', 'localhost', 'secret');
    const clusterList = await license.listClusters();
    const userClusterList = await license.listUserClusters();
    expect(clusterList.clustersList.length).toEqual(
      userClusterList.clustersList.length,
    );
    const state = await enterprise.getState();
    const activationCode = await license.getActivationCode();
    expect(state.state).toEqual(State.ACTIVE);
    expect(activationCode.activationCode).toEqual(enterpriseKey);
    await license.deleteCluster('localhost');
    await enterprise.deactivate();
  });
});
