# Changelog

<!-- <START NEW CHANGELOG ENTRY> -->
## 0.1.0-beta2

### Fixed
- Bug in unmount where we unmount then remounts the same repo
- Flakey test in mock-e2e

<!-- <END NEW CHANGELOG ENTRY> -->
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
