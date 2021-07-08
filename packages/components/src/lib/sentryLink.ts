import {ApolloLink, FetchResult, Observable} from '@apollo/client';
import {
  addBreadcrumb,
  Breadcrumb,
  configureScope,
  Scope,
  Severity,
} from '@sentry/react';

export const sentryLink = () =>
  new ApolloLink((operation, forward) => {
    const breadcrumb: Breadcrumb = {
      category: operation.query.definitions[0].kind,
      data: {
        query: operation.query.loc?.source.body || '',
      },
      level: Severity.Log,
      message: operation.operationName,
    };

    configureScope((scope: Scope) => {
      scope.setTransactionName(operation.operationName);
      scope.setFingerprint(['{{default}}', '{{transaction}}']);
    });

    return new Observable<FetchResult>((observer) => {
      const subscription = forward(operation).subscribe({
        next: (result: FetchResult) => {
          observer.next(result);
        },
        complete: () => {
          addBreadcrumb(breadcrumb);
          observer.complete();
        },
        error: (error: unknown) => {
          if (breadcrumb.data) breadcrumb.data.error = JSON.stringify(error);
          breadcrumb.level = Severity.Error;
          breadcrumb.type = 'error';
          addBreadcrumb(breadcrumb);
          return observer.error(error);
        },
      });

      return () => {
        if (subscription) subscription.unsubscribe();
      };
    });
  });
