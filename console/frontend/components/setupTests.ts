import '@testing-library/jest-dom';

import {randomBytes} from 'crypto';

import {configure} from '@testing-library/react';

configure({asyncUtilTimeout: 5000});

Object.defineProperty(window, 'crypto', {
  value: {
    getRandomValues: jest.fn((arr: number[]) => randomBytes(arr.length)),
  },
  configurable: true,
  writable: true,
});

Object.defineProperty(window.document, 'queryCommandSupported', {
  value: jest.fn(() => true),
  configurable: true,
  writable: true,
});

Object.defineProperty(window.document, 'execCommand', {
  value: jest.fn(),
  configurable: true,
  writable: true,
});

window.HTMLElement.prototype.scrollIntoView = jest.fn();
