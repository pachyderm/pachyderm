# Changelog
<!-- <START NEW CHANGELOG ENTRY> -->


# 0.6.0
New Feature: Support for Sidecar mount server

b6234e4 Add profile for sidecar (#138)
f8af207 Fix boolean logic for env vars (#139)
597d0ba Remove mount server config management from Jupyter BE (#137)
d672034 Change to Jupyter base image in Dockerfile (#135)
77ce3a1 Python 3.10 support (#132)
# 0.5.1 
d67cfc4 Auth fix - no longer crashes in the presence of inaccessible repos  (#134)
a34ce1e remove pach theme from image (#131)
0e2bc4c Add "safe to evict" to pachyderm pod (as it uses local storage) (#120)
47eaf47 Bold headings (#128)

# 0.5.0 
b42eeed Adding advanced options to the ui for server CAs (#125)
5720bb3 BE support for server cas (#124)
e78d55f Switch from umount to fusermount (#116)
11d35cd Remove python 3.6 support (#117)
d68d1da Add Ubuntu key to unblock docker-build CI job  (#121)
4cfb9c2 Make env var for setting pfs mount dir explicit (#113)
9224bd4 INT-575 (#114)
cd17f8d updating docker image in readme (#112)
2e846a9 Add instructions for TLS cert (#111)
7a10107 Move mount/unmount response generation to mount server (#92)
38885a1 update to default port (#110)
2ae61e2 Fix destroy action (#109)
0fa44b9 Bump API to v2 (#104)
ee0b1fe fix readme typo -> destory to destroy (#107)

# 0.4.0
2c9dc889 fix logo SVG, UI tweaks
2c9dc889 adding a loader component that will display till the setup function resolves
2c9dc889 use auth-related HTTP error codes in the UI

# 0.3.0-beta3
No-op: test release

# 0.3.0-beta2
53ade2e Tag docker images on tag push (#93)
e8d6432 Use config file if provided to start mount server (#91)

# 0.3.0-beta1
New Feature: Auth screen added

eb68e29 fixing tests (#89)
8d79fe7 fix for unmounted repo button (#87)
48e88fa Move updating config to python backend (#80)
7d669b0 Authentication for Mount (#81)
43de072 Return 401 from BE to FE (#84)
bab0d55 Use mount server unmount all handler (#77)
bcc43c2 Mounted Branch Status and Redesign (#78)
7cffd80 Cluster and auth backend (#72)

# 0.2.0-beta1

```
16aa82e Disable examples plugin by default (#75)
66d8cd8 Try installing wheel dynamically (#74)
6243373 Default mount server (#70)
8f00da4 Pin pach version for now (#73)
075746d Install pachctl from tarball based on any pachyderm git hash (#69)
072521d File browser chores (#64)
eb05fad Add docs (#65)
9b4c38c Pulumi preview (#59)
6b0ba49 Fixing issue where open api is getting called on mount or unmount (#63)
de444aa Docker hotfix (#62)
fe3b344 INT-483 Improve HTTP errors in handlers (#61)
705d4ca INT-487 Fix stale pypach client (#60)
7fa6d75 Build user dockerfile  (#58)
cffa0b8 [INT-454] feat(e2e): run cypress tests against real server (#53)
efed670 Remove Dockerfile (#56)
7b5d55c feat(mount): add analytics for mount/unmount (#51)
dd95119 [skip ci] Update local dev container workflow (#54)
efd7956 Modify handling of response objects to fix errors (#55)
9c6a5b9 [skip ci] Update RELEASE.md (#52)
7f821c9 Update release instructions (#36)
```

## 0.1.0-beta3

### Added
- Improved data polling in ui
### Fixed
- Bug in ui when repo has no branches

<!-- <END NEW CHANGELOG ENTRY> -->
## 0.1.0-beta2

### Fixed
- Bug in unmount where we unmount then remounts the same repo
- Flakey test in mock-e2e

## 0.1.0-beta

### Added
- Logging in the jupyterlab_pachyderm server extension
- MountServer client interface for communicating with the `pachctl mount-server`
- async request handlers

### Fixed
- Bug where deleted repos in Pachyderm still persist in list repo response

## 0.1.0-alpha3

### Added
- Examples launcher
- Mount plugin
- API to expose files in /pfs
## 0.1.0-alpha2

### Added
- Python implementation of mount backend service

## 0.1.0-alpha

### Added
- frontend help plugin
- telemetry
- Mock backend API
