interface CreateUserResponse {
  body: {
    user_id: string;
  };
}
declare namespace Cypress {
  interface Chainable {
    login(): Chainable<any>;
    logout(): Chainable<any>;
    createWorkspace(workspaceName: string): Chainable<any>;
    deleteWorkspace(workspaceName: string): Chainable<any>;
    changeOrg(orgName: string): Chainable<any>;
    openDash(workspaceName: string): Chainable<any>;
    visitDash(): Chainable<any>;
  }
}
