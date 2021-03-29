import {useLoggedInQuery} from '@dash-frontend/generated/hooks';

const useLoggedIn = () => {
  const {data} = useLoggedInQuery();

  return Boolean(data?.loggedIn);
};

export default useLoggedIn;
