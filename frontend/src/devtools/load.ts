const setAccount = async (accountId: string) => {
  await fetch(`http://localhost:30658/set-account?id=${accountId}`, {
    mode: 'cors',
  });

  localStorage.removeItem('auth-token');
  localStorage.removeItem('id-token');

  window.location.reload();
};

const setProjectTutorial = (projectId: string, tutorial: string) => {
  const key = 'active_tutorial';
  const currentSettings = JSON.parse(
    localStorage.getItem(`pachyderm-console-${projectId}`) || '{}',
  );
  localStorage.setItem(
    `pachyderm-console-${projectId}`,
    JSON.stringify({
      ...currentSettings,
      [key]: tutorial,
    }),
  );

  window.location.reload();
};

const load = () => {
  (window as unknown as Record<string, unknown>).devtools = {
    setAccount,
    setProjectTutorial,
  };
};

export default load;
