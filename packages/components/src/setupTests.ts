import '@testing-library/jest-dom';
import '@testing-library/jest-dom/extend-expect';

import {randomBytes} from 'crypto';

Object.defineProperty(window, 'crypto', {
  value: {
    getRandomValues: jest.fn((arr: number[]) => randomBytes(arr.length)),
  },
  configurable: true,
  writable: true,
});

Object.defineProperty(window, 'location', {
  configurable: true,
  value: {...window.location},
  writable: true,
});
