import React, {useState} from 'react';

import {Pager, SimplePager as SimplePagerComponent} from './Pager';

export default {title: 'Pager'};

export const Default = () => {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(5);

  const content = Array.from(Array(103).keys());

  return (
    <>
      {content
        .slice((page - 1) * pageSize, page * pageSize)
        .map((i) => i + ' ')}
      <Pager
        page={page}
        pageCount={Math.ceil(content.length / pageSize)}
        updatePage={setPage}
        pageSizes={[5, 10, 20]}
        updatePageSize={setPageSize}
        pageSize={pageSize}
      />
    </>
  );
};

export const NoPageCount = () => {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(5);

  const content = Array.from(Array(103).keys());
  const pageContent = content
    .slice((page - 1) * pageSize, page * pageSize)
    .map((i) => i + ' ');

  return (
    <>
      {pageContent}
      <Pager
        page={page}
        updatePage={setPage}
        pageSizes={[5, 10, 20]}
        nextPageDisabled={pageContent.length < pageSize}
        updatePageSize={setPageSize}
        pageSize={pageSize}
      />
    </>
  );
};

export const SimplePager = () => {
  const [page, setPage] = useState(1);
  const pageSize = 25;

  const content = Array.from(Array(103).keys());

  return (
    <>
      {content
        .slice((page - 1) * pageSize, page * pageSize)
        .map((i) => i + ' ')}
      <SimplePagerComponent
        page={page}
        pageCount={Math.ceil(content.length / pageSize)}
        updatePage={setPage}
        pageSize={pageSize}
        contentLength={content.length}
        elementName="Job"
      />
    </>
  );
};
