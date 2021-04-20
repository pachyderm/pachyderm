# This is the make version of govars.sh. It defines several common env variables
# used by Pachyderm Makefiles (GOPATH, GOBIN, and PACHCTL), both so that other
# Makefiles can be simplified by assuming these variables are set (after
# including this file) and so that those scripts work correctly on a variety of
# different platforms (Linux, Darwin, Windows) and in a variety of different
# workflows. All variables assigned here use ?= so that they can be overridden
# by environment variables/other Makefiles

ifeq ($(OS),Windows_NT)
  export GOPATH ?= $(shell cygpath -u $(shell go env GOPATH))
else
  export GOPATH ?= $(shell go env GOPATH)
endif

# Set GOBIN based on GOPATH
# TODO(msteffen) As of Apr. 2021, 'go env GOBIN' doesn't always return a path,
# but if that changes this should probably use 'go env GOBIN'.
export GOBIN ?= $(GOPATH)/bin

# Set PACHCTL based on GOBIN (want compiled pachctl to override system
# pachctl, if any)
export PACHCTL ?= $(GOBIN)/pachctl
