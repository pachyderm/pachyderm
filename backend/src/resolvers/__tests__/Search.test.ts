import {GET_SEARCH_RESULTS_QUERY} from '@dash-frontend/queries/GetSearchResultsQuery';

import {executeQuery, mockServer} from '@dash-backend/testHelpers';
import {SearchResultsQuery} from '@graphqlTypes';

describe('Search resolver', () => {
  const projectId = '1';
  it('should return no results when empty query string is provided', async () => {
    const {data, errors = []} = await executeQuery<SearchResultsQuery>(
      GET_SEARCH_RESULTS_QUERY,
      {
        args: {query: '', projectId},
      },
    );
    expect(errors).toHaveLength(0);
    const searchResults = data?.searchResults;
    expect(searchResults?.pipelines).toHaveLength(0);
    expect(searchResults?.repos).toHaveLength(0);
    expect(searchResults?.jobSet).toBeNull();
  });

  describe('Pipeline search', () => {
    it('should return pipelines when a name match is found', async () => {
      const {data, errors = []} = await executeQuery<SearchResultsQuery>(
        GET_SEARCH_RESULTS_QUERY,
        {
          args: {query: 'mon', projectId},
        },
      );

      expect(errors).toHaveLength(0);
      const searchResults = data?.searchResults;
      expect(searchResults?.pipelines).toHaveLength(1);
      expect(searchResults?.pipelines[0]?.name).toBe('montage');
    });

    it('should return no pipelines when query does not match any pipeline names', async () => {
      const {data, errors = []} = await executeQuery<SearchResultsQuery>(
        GET_SEARCH_RESULTS_QUERY,
        {
          args: {query: 'egg', projectId},
        },
      );
      expect(errors).toHaveLength(0);
      const searchResults = data?.searchResults;
      expect(searchResults?.pipelines).toHaveLength(0);
    });

    it('should not return pipelines that the user does not have access to', async () => {
      mockServer.setAccount('2');

      const {data, errors = []} = await executeQuery<SearchResultsQuery>(
        GET_SEARCH_RESULTS_QUERY,
        {
          args: {query: 'montage', projectId},
        },
      );

      expect(errors).toHaveLength(0);
      const searchResults = data?.searchResults;
      expect(searchResults?.pipelines).toHaveLength(0);
    });

    it('should return pipelines with matching jobset if globalId filter is set', async () => {
      const {data, errors = []} = await executeQuery<SearchResultsQuery>(
        GET_SEARCH_RESULTS_QUERY,
        {
          args: {
            query: 'like',
            projectId: '2',
            globalIdFilter: '23b9af7d5d4343219bc8e02ff4acd33a',
          },
        },
      );
      expect(errors).toHaveLength(0);
      const searchResults = data?.searchResults;
      expect(searchResults?.pipelines).toHaveLength(1);
      expect(searchResults?.pipelines[0]?.name).toBe('likelihoods');
    });

    it('should not return pipelines without matching jobset if globalId filter is set', async () => {
      const {data, errors = []} = await executeQuery<SearchResultsQuery>(
        GET_SEARCH_RESULTS_QUERY,
        {
          args: {
            query: 'sel',
            projectId: '2',
            globalIdFilter: '23b9af7d5d4343219bc8e02ff4acd33a',
          },
        },
      );
      expect(errors).toHaveLength(0);
      const searchResults = data?.searchResults;
      expect(searchResults?.pipelines).toHaveLength(0);
    });
  });

  describe('Job id search', () => {
    it('should return job when a job id is matched', async () => {
      const {data, errors = []} = await executeQuery<SearchResultsQuery>(
        GET_SEARCH_RESULTS_QUERY,
        {
          args: {query: '23b9af7d5d4343219bc8e02ff44cd55a', projectId},
        },
      );
      expect(errors).toHaveLength(0);
      const searchResults = data?.searchResults;
      expect(searchResults?.jobSet?.id).toBe(
        '23b9af7d5d4343219bc8e02ff44cd55a',
      );
    });

    it('should not return job when a job id is not matched', async () => {
      const {data, errors = []} = await executeQuery<SearchResultsQuery>(
        GET_SEARCH_RESULTS_QUERY,
        {
          args: {query: '23b9af7d5d4343219bc8e02ff44cd33a', projectId},
        },
      );
      expect(errors).toHaveLength(0);
      const searchResults = data?.searchResults;
      expect(searchResults?.jobSet).toBeNull();
    });
  });

  describe('Repo search', () => {
    it('should return repo when a name is matched', async () => {
      const {data, errors = []} = await executeQuery<SearchResultsQuery>(
        GET_SEARCH_RESULTS_QUERY,
        {
          args: {query: 'edg', projectId},
        },
      );
      expect(errors).toHaveLength(0);
      const searchResults = data?.searchResults;
      expect(searchResults?.repos).toHaveLength(1);
      expect(searchResults?.repos[0]?.name).toBe('edges');
    });

    it('should not return repo when a name is matched', async () => {
      const {data, errors = []} = await executeQuery<SearchResultsQuery>(
        GET_SEARCH_RESULTS_QUERY,
        {
          args: {query: 'lemon', projectId},
        },
      );
      expect(errors).toHaveLength(0);
      const searchResults = data?.searchResults;
      expect(searchResults?.repos).toHaveLength(0);
    });

    it('should not return repos that the user does not have access to', async () => {
      mockServer.setAccount('2');

      const {data, errors = []} = await executeQuery<SearchResultsQuery>(
        GET_SEARCH_RESULTS_QUERY,
        {
          args: {query: 'montage', projectId},
        },
      );
      expect(errors).toHaveLength(0);
      const searchResults = data?.searchResults;
      expect(searchResults?.repos).toHaveLength(0);
    });

    it('should return repos with matching jobset if globalId filter is set', async () => {
      const {data, errors = []} = await executeQuery<SearchResultsQuery>(
        GET_SEARCH_RESULTS_QUERY,
        {
          args: {
            query: 'like',
            projectId: '2',
            globalIdFilter: '23b9af7d5d4343219bc8e02ff4acd33a',
          },
        },
      );
      expect(errors).toHaveLength(0);
      const searchResults = data?.searchResults;
      expect(searchResults?.repos).toHaveLength(1);
      expect(searchResults?.repos[0]?.name).toBe('likelihoods');
    });

    it('should not return repos without matching jobset if globalId filter is set', async () => {
      const {data, errors = []} = await executeQuery<SearchResultsQuery>(
        GET_SEARCH_RESULTS_QUERY,
        {
          args: {
            query: 'sel',
            projectId: '2',
            globalIdFilter: '23b9af7d5d4343219bc8e02ff4acd33a',
          },
        },
      );
      expect(errors).toHaveLength(0);
      const searchResults = data?.searchResults;
      expect(searchResults?.repos).toHaveLength(0);
    });
  });
});
