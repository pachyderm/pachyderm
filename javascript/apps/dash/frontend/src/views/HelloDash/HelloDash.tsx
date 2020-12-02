import {useQuery, gql} from '@apollo/client';
import React from 'react';

import styles from './HelloDash.module.css';

const GET_HELLO_QUERY = gql`
  query hello {
    hello {
      id
      message
    }
  }
`;

const HelloDash: React.FC = () => {
  const {data} = useQuery(GET_HELLO_QUERY);

  return <h1 className={styles.base}>{data?.hello.message || 'Loading...'}</h1>;
};

export default HelloDash;
