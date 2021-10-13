import React from 'react';

import {Card} from './';

/* eslint-disable-next-line import/no-anonymous-default-export */
export default {title: 'Card'};

export const Default = () => {
  return (
    <Card autoHeight title="Monthly Invoices">
      Some info
    </Card>
  );
};
