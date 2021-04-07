export const useWorkspace = () => {
  const workspaceName = window.localStorage.getItem('workspaceName');

  return {
    workspaceName,
  };
};
