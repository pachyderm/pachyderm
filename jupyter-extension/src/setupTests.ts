import '@testing-library/jest-dom';
import {enableFetchMocks} from 'jest-fetch-mock';

Object.defineProperty(window, 'DragEvent', {
  value: class DragEvent {},
});

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation((query) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(), // deprecated
    removeListener: jest.fn(), // deprecated
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
});

// eslint-disable-next-line @typescript-eslint/no-var-requires
globalThis.crypto ??= require('node:crypto').webcrypto;

enableFetchMocks();
