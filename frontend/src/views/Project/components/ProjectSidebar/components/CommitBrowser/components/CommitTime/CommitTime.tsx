import {Commit} from '@graphqlTypes';
import formatDistanceToNow from 'date-fns/formatDistanceToNow';
import React from 'react';

type CommitTimeProps = {
  commit: Pick<Commit, 'started' | 'finished'>;
};

const CommitTime: React.FC<CommitTimeProps> = ({commit}) => {
  if (commit.finished) {
    return (
      <>{`Committed
                  ${formatDistanceToNow(commit.finished * 1000, {
                    addSuffix: true,
                  })}`}</>
    );
  } else if (commit.started) {
    return (
      <>{`Commit started ${formatDistanceToNow(commit.started * 1000, {
        addSuffix: true,
      })}`}</>
    );
  } else return null;
};

export default CommitTime;
