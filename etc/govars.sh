#!/bin/sh
#
# This is the bash version of govars.mk. It defines several common env variables
# used by Pachyderm scripts (GOPATH, GOBIN, and PACHCTL), both so that other
# scripts can be simplified by assuming these variables are set (after sourcing
# this file) and so that those scripts work correctly on a variety of different
# platforms (Linux, Darwin) and in a variety of different workflows.
#
# Because this script is sourced, it cannot assume bash and should be as
# POSIX-compliant as possible.

if test -z "${GOPATH}"
then
  GOPATH=$(go env GOPATH)
  export GOPATH # make linter happy SC2155
fi

if test -z "${GOBIN}"
then
  # Set GOBIN based on GOPATH (necessary on Windows)
  # TODO(msteffen) would it be better to use 'go env GOBIN'?
  GOBIN="${GOPATH}/bin"
  export GOBIN # make linter happy SC2155
fi

if test -z "${PACHCTL}"
then
  # Set PACHCTL based on GOBIN (want compiled pachctl to override system
  # pachctl, if any)
  PACHCTL="${GOBIN}/pachctl"
  export PACHCTL # make linter happy SC2155
fi
