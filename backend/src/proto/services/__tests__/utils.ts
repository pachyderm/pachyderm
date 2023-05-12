import client from '../../client';

export const activateEnterprise = async (authToken?: string) => {
  const enterpriseKey = process.env.PACHYDERM_ENTERPRISE_KEY || '';
  const pachClient = client({
    ssl: false,
    pachdAddress: 'localhost:30650',
    authToken,
  });
  const enterprise = pachClient.enterprise();
  const license = pachClient.license();
  await license.activateLicense(enterpriseKey);
  await license.addCluster('localhost', 'localhost:1650', 'secret');
  await license.updateCluster('localhost', 'localhost:1650', 'localhost:16650');
  await enterprise.activate('localhost:1650', 'localhost', 'secret');

  return pachClient;
};

export const deactivateEnterprise = async (authToken?: string) => {
  const pachClient = client({
    ssl: false,
    pachdAddress: 'localhost:30650',
    authToken,
  });
  const enterprise = pachClient.enterprise();
  const license = pachClient.license();
  await license.deleteCluster('localhost');
  await enterprise.deactivate();
  await license.deleteAll();

  return pachClient;
};

export const deactivateAuthEnterprise = async (authToken?: string) => {
  const pachClient = client({
    ssl: false,
    pachdAddress: 'localhost:30650',
    authToken,
  });
  const auth = pachClient.auth();
  const enterprise = pachClient.enterprise();
  const license = pachClient.license();

  await auth.deactivate();
  await license.deleteCluster('localhost');
  await enterprise.deactivate();
  await license.deleteAll();

  return pachClient;
};
