import apiClientRequestWrapper from '../../client';

export const AUTH_TOKEN = 'pizza';

export const getTestPachClient = (authToken?: string) => {
  return apiClientRequestWrapper({
    authToken: authToken ?? AUTH_TOKEN,
  });
};

export const activateEnterprise = async () => {
  const enterpriseKey = process.env.PACHYDERM_ENTERPRISE_KEY || '';
  const pachClient = getTestPachClient();

  const enterprise = pachClient.enterprise;
  const license = pachClient.license;

  await license.activateLicense(enterpriseKey);
  await license.addCluster('localhost', 'localhost:1650', 'secret');
  await license.updateCluster('localhost', 'localhost:1650', 'localhost:16650');
  await enterprise.activate('localhost:1650', 'localhost', 'secret');

  return pachClient;
};

interface deactivateEnterpriseArgs {
  deactivateAuth?: boolean;
}

export const deactivateEnterprise = async ({
  deactivateAuth = false,
}: deactivateEnterpriseArgs = {}) => {
  const pachClient = getTestPachClient();

  const enterprise = pachClient.enterprise;
  const license = pachClient.license;

  if (deactivateAuth) await pachClient.auth.deactivate();

  await license.deleteCluster('localhost');
  await enterprise.deactivate();
  await license.deleteAll();

  return pachClient;
};

export const activateAuth = async (
  pachClient: ReturnType<typeof apiClientRequestWrapper>,
) => {
  return await pachClient.auth.activate(AUTH_TOKEN);
};
