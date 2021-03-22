import {createOperation} from '@dash-backend/testHelpers';
import {SearchResults} from '@graphqlTypes';

describe('Search resolver', () => {
  describe('Pipeline search', () => {
    it('should return pipelines when a name match is found', async () => {
      const {data} = await createOperation<{
        searchResults: SearchResults;
      }>(`
      query {
        searchResults(query: "mon") {
          pipelines{
            name
            description
          }
        }
      }
    `);
      const searchResults = data?.searchResults;
      expect(searchResults?.pipelines?.length).toBe(1);
      expect(searchResults?.pipelines[0]?.name).toBe('montage');
      expect(searchResults?.pipelines[0]?.description).toBe(
        'Not my favorite pipeline',
      );
    });

    it('should return pipelines when a description match is found', async () => {
      const {data} = await createOperation<{
        searchResults: SearchResults;
      }>(`
      query {
        searchResults(query: "cool") {
          pipelines{
            name
            description
          }
        }
      }
    `);
      const searchResults = data?.searchResults;
      expect(searchResults?.pipelines?.length).toBe(1);
      expect(searchResults?.pipelines[0]?.name).toBe('edges');
      expect(searchResults?.pipelines[0]?.description).toBe(
        'Very cool edges description',
      );
    });

    it('should return no pipelines when query does not match any pipeline names or descriptions', async () => {
      const {data} = await createOperation<{
        searchResults: SearchResults;
      }>(`
      query {
        searchResults(query: "egg") {
          pipelines{
            name
          }
        }
      }
    `);
      const searchResults = data?.searchResults;
      expect(searchResults?.pipelines?.length).toBe(0);
    });

    it('should return no pipelines when empty query string is provided', async () => {
      const {data} = await createOperation<{
        searchResults: SearchResults;
      }>(`
      query {
        searchResults(query: "") {
          pipelines{
            name
          }
        }
      }
    `);
      const searchResults = data?.searchResults;
      expect(searchResults?.pipelines?.length).toBe(0);
    });
  });
});
