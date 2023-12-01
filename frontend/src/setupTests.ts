import 'cross-fetch';
import '@testing-library/jest-dom';

import {randomBytes} from 'crypto';

import {configure} from '@testing-library/react';
import {enableFetchMocks} from 'jest-fetch-mock';

configure({asyncUtilTimeout: 5000});

enableFetchMocks();

Object.defineProperty(window, 'crypto', {
  value: {
    getRandomValues: jest.fn((arr: number[]) => randomBytes(arr.length)),
  },
  configurable: true,
  writable: true,
});

class MockObserver implements IntersectionObserver {
  readonly root: Element | null = null;
  readonly rootMargin: string = '';
  readonly thresholds: ReadonlyArray<number> = [];
  disconnect: () => void = () => null;
  observe: (target: Element) => void = () => null;
  takeRecords: () => IntersectionObserverEntry[] = () => [];
  unobserve: (target: Element) => void = () => null;
}

Object.defineProperty(window, 'IntersectionObserver', {
  writable: true,
  configurable: true,
  value: MockObserver,
});

Object.defineProperty(global, 'IntersectionObserver', {
  writable: true,
  configurable: true,
  value: MockObserver,
});

Object.defineProperty(global, 'ResizeObserver', {
  writable: true,
  configurable: true,
  value: MockObserver,
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

Object.assign(navigator, {
  clipboard: {
    writeText: jest.fn(),
  },
});

// This method needs an implementation due to CodeMirror
// https://github.com/jsdom/jsdom/pull/3533
// https://github.com/jsdom/jsdom/issues/3002
// https://github.com/jsdom/jsdom/issues/2751
Object.defineProperty(window.Range.prototype, 'getClientRects', {
  value: jest.fn(() => [
    {
      bottom: 0,
      height: 0,
      left: 0,
      right: 0,
      top: 0,
      width: 0,
    },
  ]),
  configurable: true,
  writable: true,
});

window.HTMLElement.prototype.scrollIntoView = jest.fn();

// DISABLE POLLING FOR MSW TESTS
process.env.REACT_APP_POLLING = '0';

jest.setTimeout(10_000);
