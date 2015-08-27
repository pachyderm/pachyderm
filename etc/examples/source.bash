REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

run() {
  echo $@ >&2
  $@
}

run_make() {
  run make -C "${REPO_DIR}" $@
}

do_pfs() {
  if [ -z ${PFS_CLIENT_EXAMPLE} ]; then
    run_make launch-pfsd
    run docker run -e PFS_CLIENT_EXAMPLE=1 --link pachyderm_pfsd_1 pachyderm_compile bash "etc/examples/$(basename "${1}")"
    run_make docker-clean-launch
  else
    run_make install
    ${2}
  fi
}

do_all() {
  if [ -z ${PPS_CLIENT_EXAMPLE} ]; then
    run_make launch
    run docker run -e PPS_CLIENT_EXAMPLE=1 --link pachyderm_pfsd_1 --link pachyderm_ppsd_1 pachyderm_compile bash "etc/examples/$(basename "${1}")"
    run_make docker-clean-launch
  else
    run_make install
    ${2}
  fi
}
