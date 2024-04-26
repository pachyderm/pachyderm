import {chromium, Browser, BrowserContext, errors} from 'playwright';
import {iToSeconds, pickRandomJourney, wait} from './utils';
import {
  ADDED_SLOWMOTION,
  BASE_URL,
  EXPECT_TIMEOUT,
  HEADLESS,
  INITIAL_PAGELOAD_SECONDS_MULTIPLIER,
} from './constants';
import {journeys, loginWithMockAdmin} from './userJourneys';
import {statsLogger} from './logger';

let shutdownFlag = false;

export const enableShutDownFlag = () => {
  shutdownFlag = true;
};

// Startup
const startBrowser = async () =>
  chromium.launch({slowMo: ADDED_SLOWMOTION, headless: HEADLESS});

export const startBrowsers = async (count: number) => {
  console.log('Initializing', count, 'browsers...');
  const browsers = await Promise.all<Browser>(
    Array.from({length: count}, () => startBrowser()),
  );
  browsers.forEach((browser, index) =>
    browser.on('disconnected', () =>
      console.log('Browser', index, 'disconnected.'),
    ),
  );

  console.log(count, 'browsers initialized!');
  return browsers;
};

const createContexts = async (
  browsers: Browser[],
): Promise<BrowserContext[]> => {
  const contexts: BrowserContext[] = [];
  let i = 1;
  for (const browser of browsers) {
    const context = await browser.newContext({baseURL: BASE_URL});
    contexts.push(context);
    const index = i;
    context.on('close', () => console.log('context', index, 'closed.'));
    i++;
  }
  return contexts;
};

const configureContexts = async (
  contexts: BrowserContext[],
  authEnabled?: boolean,
): Promise<void> => {
  for (const context of contexts) {
    context.setDefaultNavigationTimeout(EXPECT_TIMEOUT);
    context.setDefaultTimeout(EXPECT_TIMEOUT);

    if (authEnabled) {
      const page = await context.newPage();
      await loginWithMockAdmin(page);
    }
  }
};

export const setupContexts = async (
  browsers: Browser[],
  authEnabled?: boolean,
): Promise<BrowserContext[]> => {
  const contexts = await createContexts(browsers);
  await configureContexts(contexts, authEnabled);
  return contexts;
};

export const tearDown = async (thing: Browser | BrowserContext) => {
  await thing.close();
};

// Do stuff
export const runJourney = async (context: BrowserContext, index: number) => {
  const startDelay = iToSeconds(index);
  await wait(startDelay * INITIAL_PAGELOAD_SECONDS_MULTIPLIER);

  const logger = (message: string) => console.log(`${index}: ${message}`);

  // eslint-disable-next-line no-constant-condition
  while (true) {
    const page = await context.newPage();

    page.route(new RegExp(BASE_URL + 'api'), (route) => {
      const request = route.request();
      const timings = request.timing();
      const url = request.url();

      statsLogger.writeTiming(url, timings);

      return route.continue();
    });

    const journey = pickRandomJourney(journeys);

    try {
      logger(`Starting journey: '${journey.name}'`);
      await journey.journey(page);
      logger(`Finished journey: ${journey.name}`);
      statsLogger.writeSuccess(journey.name, index);
    } catch (e) {
      if (e instanceof errors.TimeoutError) {
        logger(`I timed out! Journey: '${journey.name}'`);
        statsLogger.writeError(journey.name, index, e);
      } else if (shutdownFlag) {
        break;
      } else {
        logger(`Encountered an error! Journey: ${journey.name}`);
        statsLogger.writeError(journey.name, index, e as Error);
      }

      console.error(e);
    } finally {
      await page.close();
    }
  }
};
