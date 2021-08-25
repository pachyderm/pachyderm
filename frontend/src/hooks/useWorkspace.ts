export const useWorkspace = () => {
  const workspaceName = window.localStorage.getItem('workspaceName');
  const pachVersion = window.localStorage.getItem('pachVersion');
  const pachdAddress = window.localStorage.getItem('pachdAddress');
  const hasConnectInfo = Boolean(pachVersion && pachdAddress);

  return {
    workspaceName,
    pachVersion,
    pachdAddress,
    hasConnectInfo,
  };
};
