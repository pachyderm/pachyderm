import * as React from 'react';

const SvgStatusWarning = (props) => (
  <svg
    width={20}
    height={20}
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 16 16"
    {...props}
  >
    <g fill="none" fillRule="evenodd">
      <path d="M0 0h16v16H0z" />
      <path
        d="M8 0a8 8 0 1 1 0 16A8 8 0 0 1 8 0Zm0 11.65a1.2 1.2 0 1 0 0 2.4 1.2 1.2 0 0 0 0-2.4Zm0-9.7a1.2 1.2 0 0 0-1.2 1.2v5.6a1.2 1.2 0 1 0 2.4 0v-5.6A1.2 1.2 0 0 0 8 1.95Z"
        fill="#BF444F"
      />
    </g>
  </svg>
);

export default SvgStatusWarning;
