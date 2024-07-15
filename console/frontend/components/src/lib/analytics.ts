/* eslint-disable @typescript-eslint/naming-convention */
import {captureException} from '@sentry/react';
import Cookies from 'js-cookie';
import debounce from 'lodash/debounce';
import each from 'lodash/each';
import isEmpty from 'lodash/isEmpty';
import omitBy from 'lodash/omitBy';

export type EventType = 'identify' | 'page' | 'track';

/* eslint-disable @typescript-eslint/no-explicit-any */
type Track = (...args: any[]) => void;
type Page = (...args: any[]) => void;
type Identify = (...args: any[]) => void;
/* eslint-enable @typescript-eslint/no-explicit-any */

enum QueryStringCookies {
  utm_campaign = 'utm_campaign',
  utm_source = 'utm_source',
  utm_medium = 'utm_medium',
  utm_content = 'utm_content',
  utm_term = 'utm_term',
}

export const onTitleChange = (cb: (title: string) => void) => {
  const target = document.querySelector('title');
  const observer = new MutationObserver((mutations) => {
    mutations.forEach((mutation) => {
      if (mutation.target === target) {
        cb(document.title);
      }
    });
  });

  if (target) observer.observe(target, {childList: true});

  return observer;
};

export const CLICK_TIMEOUT = 500;
export const COOKIE_EXPIRATION_DAYS = 365;

export const getTrackingCookies = () => {
  return {
    latest_utm_source: Cookies.get('latest_utm_source'),
    latest_utm_campaign: Cookies.get('latest_utm_campaign'),
    latest_utm_medium: Cookies.get('latest_utm_medium'),
    latest_utm_content: Cookies.get('latest_utm_content'),
    latest_utm_term: Cookies.get('latest_utm_term'),
    source_utm_campaign: Cookies.get('source_utm_campaign'),
    source_utm_source: Cookies.get('source_utm_source'),
    source_utm_medium: Cookies.get('source_utm_medium'),
    source_utm_content: Cookies.get('source_utm_content'),
    source_utm_term: Cookies.get('source_utm_term'),
  };
};

export const captureTrackingCookies = () => {
  const search = new URLSearchParams(window.location.search);

  each(QueryStringCookies, (searchParam: string) => {
    const searchValue = search.get(searchParam);

    if (searchValue && !Cookies.get(`source_${searchParam}`)) {
      Cookies.set(`source_${searchParam}`, searchValue, {
        expires: COOKIE_EXPIRATION_DAYS,
      });
    }

    if (searchValue) {
      Cookies.set(`latest_${searchParam}`, searchValue, {
        expires: COOKIE_EXPIRATION_DAYS,
      });
    }
  });
};

export const fireClick = (clickId: string, track: Track) => {
  try {
    track('click', {clickId});
  } catch (err) {
    captureException(
      `[Analytics Error]: Operation: track, Event: click, ID: ${clickId}, ${err}`,
    );
  }
};

export const fireIdentify = (
  authId: string,
  authEmail: string,
  authCreatedAt: number,
  identify: Identify,
  track: Track,
) => {
  try {
    const trackingCookies = getTrackingCookies();
    const createdAt = new Date(authCreatedAt * 1000).toISOString();

    identify(authId, {
      email: authEmail,
      created_at: createdAt,
      ...trackingCookies,
    });

    track('authenticated', {
      authCreatedAt: createdAt,
      authEmail,
      authId,
      ...trackingCookies,
    });
  } catch (err) {
    captureException(
      `[Analytics Error]: Operation: track, Event: identify, ID: ${authId}, ${err}`,
    );
  }
};

export const fireClusterInfo = (clusterId: string, track: Track) => {
  try {
    track('cluster_info', {clusterId});
  } catch (err) {
    captureException(
      `[Analytics Error]: Operation: track, Event: clusterId, ID: ${clusterId}, ${err}`,
    );
  }
};

export const firePageView = (page: Page) => {
  try {
    page();
  } catch (err) {
    captureException(`[Analytics Error]: Operation: page, ${err}`);
  }
};

export const fireUTM = (track: Track) => {
  const traits = omitBy(getTrackingCookies(), (_, k) => k.startsWith('source'));

  if (!isEmpty(traits)) {
    try {
      track('UTM', {context: {traits}});
    } catch (err) {
      captureException(
        `[Analytics Error]: Operation: track, Event: UTM, ${err}`,
      );
    }
  }
};

export const initClickTracker = (track: Track) => {
  window.document.addEventListener(
    'click',
    debounce(
      (evt) => {
        const element = evt.target as HTMLElement;
        const testId = element.getAttribute('data-testid');

        if (testId) {
          fireClick(testId, track);
        }
      },
      CLICK_TIMEOUT,
      {
        leading: true,
        trailing: false,
      },
    ),
  );
};

export const initPageTracker = (track: Track) => {
  return onTitleChange(() => firePageView(track));
};
