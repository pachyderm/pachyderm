helmlib is a utility library for deploying Pachyderm into a Kubernetes cluster
using its helm chart. It reduces the complexity of using the helm chart
directly by deploying Pachyderm in one of a few well-known testing
configurations.

Currently, this code is only used as part of the `src/testing/pachdev` CLI tool
(which is officially the local testing tool of Pachyderm, and is also planned
to be used, at least, by our pachyderm-sdk python tests.

This library in general, and deploy.go in particular, has a lot of code copied
from `src/internal/minikubetestenv`, which is a Go library that allows
Pachyderm's Go tests to deploy their own clusters as part of test setup.
`minikubetestenv` precedes `pachdev`, but could not be used by our python tests
or a CLI tool because it's written to be included in Go tests specifically
(taking `*testing.T` everywhere, for example). Eventually, however, these two
should be unified, such that all this code is factored out of `minikubetestenv`
and `minikubetestenv` calls into `helmlib` for deployment. This was not done
during 'pachdev' development (which is when this library was added) to limit
the scope of that project--rolling out a new developer tool is complicated
enough without having rewrite tests at the same time.

However, given that long-term vision, this README exists to catalogue the
remaining work to this end. This sort of refactoring is not on any one team's
roadmap, which is why the TODO list exists here instead of in Jira.

## TODOs
- [ ] Migrate `src/internal/testutil/auth.go` into a library in `src/testing`
  - `auth.go` is a utility library that assists with the process of deploying
    Pachyderm into a Kubernetes cluster using its helm chart. Like deploy.go,
    this library has a lot of copied code, but in this case, the code copied
    from `src/internal/testutil/auth.go`, which is a Go library that assists
    `minikubetestenv` in the same way that `auth.go` assists `deploy.go`. This
    was not done during 'pachdev' development for the same reasons as
    `deploy.go`--widespread use of `*testing.T` in `testutil/auth.go` and
    limiting the scope of developing `pachdev`.
- [ ] Determined deployment is basically not implemented in `helmlib`
  - The semantics of Determined deployment in `minikubetestenv` seem very
    specific to whatever test suite or project it was added to support. For
    example, I don't know how you'd simply deploy the most recently released
    Determined images on Dockerhub, which seems like the first use-case that
    should work. I worry that just including the current `minikubetestenv`
    implementation of deploying Determined will necessitate the addition of ton
    of weird and ultimately unnecessary startup parameters to `pachdev`,
    turning it into a mid-layer (where we're trying to pass all helm options
    through and no complexity is reduced). It might alternatively be better for
    callers to deploy Determined themselves.
- [ ] `EnterpriseServer` and `EnterpriseMember` are not supported in `helmlib`
  - The helm values set by `withEnterpriseServer` do not make a
    lot of sense to me. In particular, it overwrites a bunch of pachd service
    options, like `proxy.service.httpNodePort`, and I don't understand why
    pachd's public address needs a specific value in this configuration
    (particularly one that prevents any other pachd instances from running in
    the same cluster). We may be getting rid of the Enterprise Server soon, so
    rather than figure out how to port that code I've opted to simply comment
    out everything related to the EnterpriseServer and EnterpriseMember
    options. I'm leaving this note here in case that decision becomes an issue
    later.
- [ ] I have avoided porting `withPort` to `helmlib` and it simply avoids
    setting many port-related helm values.
  - `withPort` in `minikubetestenv` takes great pains to ensure that
    Pachyderm's various services (dex, the S3 gateway, etc) all use the same
    port inside k8s as outside k8s, setting such helm values as
    `pachd.service.apiGRPCPort`, `pachd.service.oidcPort`, and
    `pachd.service.identityPort`. I don't yet understand why this is desirable,
    nor why this is set explicitly by `minikubetestenv` instead of being an
    implementation detail of the helm chart. 
