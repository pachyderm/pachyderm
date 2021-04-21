import React from 'react';
import {BrowserRouter} from 'react-router-dom';

import {ButtonLink} from '../ButtonLink';

import Link from './';

export default {title: 'Link'};

export const Default = () => {
  return (
    <BrowserRouter>
      <Link to="/">Default Link</Link>
    </BrowserRouter>
  );
};

export const Small = () => {
  return (
    <BrowserRouter>
      <Link small to="/">
        Small Link
      </Link>
    </BrowserRouter>
  );
};

export const Inline = () => {
  return (
    <BrowserRouter>
      Click this{' '}
      <Link inline to="/">
        inline link
      </Link>
      .
    </BrowserRouter>
  );
};

export const External = () => {
  return (
    <Link externalLink to="https://google.com">
      Google
    </Link>
  );
};

export const Disabled = () => {
  return <Link>Disabled</Link>;
};

export const Button = () => {
  return <ButtonLink>Button</ButtonLink>;
};
