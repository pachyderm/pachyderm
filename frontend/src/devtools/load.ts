const setAccount = async (accountId: string) => {
  await fetch(`http://localhost:30658/set-account?id=${accountId}`, {
    mode: 'cors',
  });

  localStorage.removeItem('auth-token');
  localStorage.removeItem('id-token');

  window.location.reload();
};

const load = () => {
  (window as unknown as Record<string, unknown>).devtools = {
    setAccount,
  };
};

export default load;
