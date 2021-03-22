import {gql} from '@apollo/client';

export const GET_JOBS_QUERY = gql`
  query getJobs($args: JobQueryArgs!) {
    jobs(args: $args) {
      id
      state
      createdAt
    }
  }
`;
