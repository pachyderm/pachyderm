import {gql} from '@apollo/client/core';

import {executeQuery} from '@dash-backend/testHelpers';
import {SearchResults} from '@graphqlTypes';

describe('Search resolver', () => {
  describe('Pipeline search', () => {
    it('should return pipelines when a name match is found', async () => {
      const query = gql`
        query {
          searchResults(query: "mon") {
            pipelines {
              name
              description
            }
          }
        }
      `;

      const {data} = await executeQuery<{
        searchResults: SearchResults;
      }>(query);

      const searchResults = data?.searchResults;
      expect(searchResults?.pipelines?.length).toBe(1);
      expect(searchResults?.pipelines[0]?.name).toBe('montage');
      expect(searchResults?.pipelines[0]?.description).toBe(
        'Not my favorite pipeline',
      );
    });

    it('should return pipelines when a description match is found', async () => {
      const query = gql`
        query {
          searchResults(query: "cool") {
            pipelines {
              name
              description
            }
          }
        }
      `;

      const {data} = await executeQuery<{
        searchResults: SearchResults;
      }>(query);

      const searchResults = data?.searchResults;
      expect(searchResults?.pipelines?.length).toBe(1);
      expect(searchResults?.pipelines[0]?.name).toBe('edges');
      expect(searchResults?.pipelines[0]?.description).toBe(
        'Very cool edges description',
      );
    });

    it('should return no pipelines when query does not match any pipeline names or descriptions', async () => {
      const query = gql`
        query {
          searchResults(query: "egg") {
            pipelines {
              name
            }
          }
        }
      `;

      const {data} = await executeQuery<{
        searchResults: SearchResults;
      }>(query);

      const searchResults = data?.searchResults;
      expect(searchResults?.pipelines?.length).toBe(0);
    });

    it('should return no pipelines when empty query string is provided', async () => {
      const query = gql`
        query {
          searchResults(query: "") {
            pipelines {
              name
            }
          }
        }
      `;

      const {data} = await executeQuery<{
        searchResults: SearchResults;
      }>(query);

      const searchResults = data?.searchResults;
      expect(searchResults?.pipelines?.length).toBe(0);
    });
  });

  describe('Job id search', () => {
    it('should return job when a job id is matched', async () => {
      const query = gql`
        query {
          searchResults(query: "23b9af7d5d4343219bc8e02ff44cd55a") {
            pipelines {
              name
            }
            job {
              id
              createdAt
            }
          }
        }
      `;

      const {data} = await executeQuery<{
        searchResults: SearchResults;
      }>(query);

      const searchResults = data?.searchResults;
      expect(searchResults?.pipelines?.length).toBe(0);
      expect(searchResults?.job?.id).toBe('23b9af7d5d4343219bc8e02ff44cd55a');
      expect(searchResults?.job?.createdAt).toBe(1616533099);
    });

    it('should not return job when a job id is not matched', async () => {
      const query = gql`
        query {
          searchResults(query: "23b9af7d5d4343219bc8e02ff44cd33a") {
            pipelines {
              name
            }
            job {
              id
              createdAt
            }
          }
        }
      `;

      const {data} = await executeQuery<{
        searchResults: SearchResults;
      }>(query);

      const searchResults = data?.searchResults;
      expect(searchResults?.pipelines?.length).toBe(0);
      expect(searchResults?.job).toBe(null);
    });
  });
});
