const NETWORK_REQUESTS_PATH = Cypress.env('NETWORK_REQUESTS_PATH') || '';
const NETWORK_REQUESTS = {};

before(() => {
  if (!NETWORK_REQUESTS_PATH) return;

  Cypress.on('window:before:load', (win) => {
    const realFetch = win.fetch;

    win.fetch = async (url, ...args) => {
      const realResult = await realFetch(url, ...args);
      const json = await realResult.clone().json();
      const requestKey = `${args[0].method}:${url}`

      if (!NETWORK_REQUESTS[requestKey]) {
        NETWORK_REQUESTS[requestKey] = [];
      }

      NETWORK_REQUESTS[requestKey].push(json);

      return realResult;
    };
  });
});

after(() => {
  if (!NETWORK_REQUESTS_PATH) return;

  cy.writeFile(NETWORK_REQUESTS_PATH, JSON.stringify(NETWORK_REQUESTS, null, 2));
});
