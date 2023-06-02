import {useFileDownloadLazyQuery} from '@dash-frontend/generated/hooks';

const useFileDownload = () => {
  const [fileDownload, rest] = useFileDownloadLazyQuery();

  return {
    fileDownload,
    ...rest,
    url: rest.data?.fileDownload,
  };
};

export default useFileDownload;
