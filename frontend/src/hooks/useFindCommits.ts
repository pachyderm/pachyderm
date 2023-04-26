import {useFindCommitsLazyQuery} from '@dash-frontend/generated/hooks';

const useFindCommits = () => {
  const [findCommits, rest] = useFindCommitsLazyQuery();

  return {
    findCommits,
    ...rest,
    commits: rest.data?.findCommits.commits,
    cursor: rest.data?.findCommits.cursor,
    hasNextPage: rest.data?.findCommits.hasNextPage,
  };
};

export default useFindCommits;
