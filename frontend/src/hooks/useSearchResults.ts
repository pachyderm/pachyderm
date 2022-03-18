import {useSearchResultsQuery} from '@dash-frontend/generated/hooks';

export const useSearchResults = (
  projectId: string,
  query: string,
  limit?: number,
  globalIdFilter?: string,
) => {
  const {data, error, loading, previousData} = useSearchResultsQuery({
    variables: {
      args: {
        projectId,
        query: query || '',
        limit: limit,
        globalIdFilter,
      },
    },
  });

  return {
    error,
    searchResults: data?.searchResults || null,
    previousSearchResults: previousData,
    loading,
  };
};
