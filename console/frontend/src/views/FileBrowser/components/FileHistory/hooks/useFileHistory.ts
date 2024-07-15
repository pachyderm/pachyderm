import objectHash from 'object-hash';
import {useEffect, useMemo, useState} from 'react';

import {CommitInfo} from '@dash-frontend/api/pfs';
import {useFindCommits} from '@dash-frontend/hooks/useFindCommits';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {getStandardDateFromISOString} from '@dash-frontend/lib/dateTime';

const useFileHistory = () => {
  const {repoId, projectId, filePath, commitId} = useUrlState();
  const [commitCursors, setCommitCursors] = useState<string[]>([]);
  const [commitList, setCommitList] = useState<CommitInfo[]>([]);

  const req = useMemo(
    () => ({
      start: {
        id:
          commitCursors.length > 0
            ? commitCursors[commitCursors.length - 1]
            : commitId,
        branch: {
          //TODO: Remove branch once core fixes
        },
        repo: {name: repoId, project: {name: projectId}, type: 'user'},
      },
      filePath: `/${filePath}`,
    }),
    [commitCursors, commitId, filePath, projectId, repoId],
  );

  const {findCommits, commits, cursor, loading, error} = useFindCommits(req);

  useEffect(() => {
    if (commits && !loading) {
      setCommitList((prev) => ({...prev, [objectHash(req)]: commits}));
      if (cursor && !commitCursors.includes(cursor))
        setCommitCursors((commitCursors) => [...commitCursors, cursor]);
    }
  }, [cursor, commits, loading, req, commitCursors]);

  const flatCommits = useMemo(
    () => Object.values(commitList).flat(),
    [commitList],
  );

  const dateRange = useMemo(() => {
    return flatCommits.length !== 0
      ? `${getStandardDateFromISOString(flatCommits[0]?.started)} - ${
          flatCommits &&
          getStandardDateFromISOString(
            flatCommits[flatCommits.length - 1]?.started,
          )
        }`
      : null;
  }, [flatCommits]);

  // disable search if we have loaded all the pages possible and the last request did return a cursor
  const disableSearch =
    Object.values(commitList).length >= commitCursors.length + 1;

  return {
    findCommits,
    loading,
    commitList: flatCommits,
    dateRange,
    disableSearch,
    error,
  };
};

export default useFileHistory;
