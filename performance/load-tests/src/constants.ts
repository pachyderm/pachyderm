import {parseArgs} from 'node:util';

const {values} = parseArgs({
  options: {
    baseURL: {type: 'string'},
    testRuntime: {type: 'string'},
    browserCount: {type: 'string'},
    expectTimeout: {type: 'string'},
    addedSlowmotion: {type: 'string'},
    initialPageloadDelay: {type: 'string'},
    noHeadless: {type: 'boolean'},
    debug: {type: 'boolean'},
    auth: {type: 'boolean'},
  },
});

export const BASE_URL = String(values.baseURL || 'http://localhost:4000/');
export const TEST_RUNTIME = Number(values.testRuntime || 5 * 60 * 1000); // 5 minutes
export const BROWSER_COUNT = Number(values.browserCount || 10);

export const EXPECT_TIMEOUT = Number(values.expectTimeout || 60 * 1000); // 60 seconds
export const ADDED_SLOWMOTION = Number(values.addedSlowmotion || 1000);
/** When the browsers start hitting the server, they are staggered by 1 second each. This multiplies that time. */
export const INITIAL_PAGELOAD_SECONDS_MULTIPLIER = Number(
  values.initialPageloadDelay || 1,
);

export const HEADLESS = Boolean(values.noHeadless ? false : true);

export const DEBUG = Boolean(values.debug);
export const AUTH = Boolean(values.auth);

DEBUG &&
  console.log({
    BASE_URL,
    TEST_RUNTIME,
    BROWSER_COUNT,
    EXPECT_TIMEOUT,
    ADDED_SLOWMOTION,
    INITIAL_PAGELOAD_SECONDS_MULTIPLIER,
    HEADLESS,
    AUTH,
  });
