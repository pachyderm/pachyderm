/* eslint-disable @typescript-eslint/naming-convention */
import {captureException} from '@sentry/react';
import Cookies from 'js-cookie';
import debounce from 'lodash/debounce';
import each from 'lodash/each';
import isEmpty from 'lodash/isEmpty';
import omitBy from 'lodash/omitBy';
import {getAnonymousId, identify, page, track} from 'rudder-sdk-js';

export type EventType = 'identify' | 'page' | 'track';

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

export const fireClick = (clickId: string) => {
  try {
    track('click', {clickId});
  } catch (err) {
    captureException(
      `[Rudderstack Error]: Operation: track, Event: click, ID: ${clickId}, ${err}`,
    );
  }
};

export const fireIdentify = (
  authId: string,
  authEmail: string,
  authCreatedAt: number,
) => {
  try {
    const trackingCookies = getTrackingCookies();
    const anonymousId = getAnonymousId();
    const createdAt = new Date(authCreatedAt * 1000).toISOString();

    identify(authId, {
      anonymous_id: anonymousId,
      email: authEmail,
      hub_created_at: createdAt,
      hub_promo_code: undefined,
      hub_user_id: authId,
      ...trackingCookies,
    });

    track('authenticated', {
      authCreatedAt: createdAt,
      authEmail,
      authId,
      anonymousId,
      ...trackingCookies,
    });
  } catch (err) {
    captureException(
      `[Rudderstack Error]: Operation: track, Event: identify, ID: ${authId}, ${err}`,
    );
  }
};

export const firePageView = () => {
  try {
    page();
  } catch (err) {
    captureException(`[Rudderstack Error]: Operation: page, ${err}`);
  }
};

export const firePromoApplied = (promo: string, authId: string) => {
  try {
    identify(authId, {hub_promo_code: promo});
    track('Promo', {
      context: {
        traits: {
          hub_promo_code: promo,
        },
      },
    });
  } catch (err) {
    captureException(
      `[Rudderstack Error]: Operation: track, Event: promo, ID: ${authId}, Promo: ${promo}, ${err}`,
    );
  }
};

export const fireUTM = () => {
  const traits = omitBy(getTrackingCookies(), (_, k) => k.startsWith('source'));

  if (!isEmpty(traits)) {
    try {
      track('UTM', {context: {traits}});
    } catch (err) {
      captureException(
        `[Rudderstack Error]: Operation: track, Event: UTM, ${err}`,
      );
    }
  }
};

export const initClickTracker = () => {
  window.document.addEventListener(
    'click',
    debounce(
      (evt) => {
        const element = evt.target as HTMLElement;
        const testId = element.getAttribute('data-testid');

        if (testId) {
          fireClick(testId);
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

export const initPageTracker = () => {
  return onTitleChange(() => firePageView());
};
