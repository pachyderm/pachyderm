import {useSearchResultsQuery} from '@dash-frontend/generated/hooks';

export const useSearchResults = (
  projectId: string,
  query: string,
  limit?: number,
) => {
  const {data, error, loading, previousData} = useSearchResultsQuery({
    variables: {
      args: {
        projectId,
        query: query || '',
        limit: limit,
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
