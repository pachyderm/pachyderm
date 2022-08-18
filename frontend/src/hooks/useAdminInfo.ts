import {useGetAdminInfoQuery} from '@dash-frontend/generated/hooks';

interface useAdminInfoArgs {
  skip?: boolean;
}

const useAdminInfo = ({skip = false}: useAdminInfoArgs = {}) => {
  const {data, error, loading} = useGetAdminInfoQuery({skip});
  return {
    error,
    loading,
    clusterId: data?.adminInfo.clusterId,
  };
};

export default useAdminInfo;
