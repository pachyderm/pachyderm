import {State} from '../../proto/enterprise/enterprise_pb';

import {activateEnterprise, deactivateEnterprise} from './utils';

describe('enterprise', () => {
  afterEach(async () => {
    await deactivateEnterprise();
  });

  it('should activate enterprise', async () => {
    expect(process.env.PACHYDERM_ENTERPRISE_KEY).not.toBeFalsy();

    const pachClient = await activateEnterprise();
    const enterprise = pachClient.enterprise;
    const license = pachClient.license;

    const clusterList = await license.listClusters();
    const userClusterList = await license.listUserClusters();
    expect(clusterList.clustersList).toHaveLength(
      userClusterList.clustersList.length,
    );

    const state = await enterprise.getState();
    const activationCode = await license.getActivationCode();
    expect(state.state).toEqual(State.ACTIVE);
    expect(activationCode.activationCode).toEqual(
      process.env.PACHYDERM_ENTERPRISE_KEY,
    );
  });
});
