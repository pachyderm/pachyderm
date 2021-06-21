import {executeQuery, mockServer} from '@dash-backend/testHelpers';
import {GET_SEARCH_RESULTS_QUERY} from '@dash-frontend/queries/GetSearchResultsQuery';
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
    expect(errors.length).toBe(0);
    const searchResults = data?.searchResults;
    expect(searchResults?.pipelines?.length).toBe(0);
    expect(searchResults?.repos?.length).toBe(0);
    expect(searchResults?.jobset).toBe(null);
  });

  describe('Pipeline search', () => {
    it('should return pipelines when a name match is found', async () => {
      const {data, errors = []} = await executeQuery<SearchResultsQuery>(
        GET_SEARCH_RESULTS_QUERY,
        {
          args: {query: 'mon', projectId},
        },
      );
      expect(errors.length).toBe(0);
      const searchResults = data?.searchResults;
      expect(searchResults?.pipelines?.length).toBe(1);
      expect(searchResults?.pipelines[0]?.name).toBe('montage');
    });

    it('should return no pipelines when query does not match any pipeline names', async () => {
      const {data, errors = []} = await executeQuery<SearchResultsQuery>(
        GET_SEARCH_RESULTS_QUERY,
        {
          args: {query: 'egg', projectId},
        },
      );
      expect(errors.length).toBe(0);
      const searchResults = data?.searchResults;
      expect(searchResults?.pipelines?.length).toBe(0);
    });

    it('should not return pipelines that the user does not have access to', async () => {
      mockServer.setAccount('2');

      const {data, errors = []} = await executeQuery<SearchResultsQuery>(
        GET_SEARCH_RESULTS_QUERY,
        {
          args: {query: 'montage', projectId},
        },
      );

      expect(errors.length).toBe(0);
      const searchResults = data?.searchResults;
      expect(searchResults?.pipelines?.length).toBe(0);
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
      expect(errors.length).toBe(0);
      const searchResults = data?.searchResults;
      expect(searchResults?.jobset?.id).toBe(
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
      expect(errors.length).toBe(0);
      const searchResults = data?.searchResults;
      expect(searchResults?.jobset).toBe(null);
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
      expect(errors.length).toBe(0);
      const searchResults = data?.searchResults;
      expect(searchResults?.repos?.length).toBe(1);
      expect(searchResults?.repos[0]?.name).toBe('edges');
    });

    it('should not return repo when a name is matched', async () => {
      const {data, errors = []} = await executeQuery<SearchResultsQuery>(
        GET_SEARCH_RESULTS_QUERY,
        {
          args: {query: 'lemon', projectId},
        },
      );
      expect(errors.length).toBe(0);
      const searchResults = data?.searchResults;
      expect(searchResults?.repos?.length).toBe(0);
    });

    it('should not return repos that the user does not have access to', async () => {
      mockServer.setAccount('2');

      const {data, errors = []} = await executeQuery<SearchResultsQuery>(
        GET_SEARCH_RESULTS_QUERY,
        {
          args: {query: 'montage', projectId},
        },
      );
      expect(errors.length).toBe(0);
      const searchResults = data?.searchResults;
      expect(searchResults?.repos?.length).toBe(0);
    });
  });
});
