#!/bin/bash
################################################################################
# This script is what I (msteffen) have been using to build, run, and test
# Pachyderm more or less since I joined (2016). It's similar to restart.py in
# the parent directory. Some parts may be out of date (notably, it doesn't use
# minikube tunnel--for some reason, that didn't work the last time I tried it),
# but as of Feb. 2023, it still basically works.
################################################################################

# Get the dir in which this script resides (git-root/etc/contributing), working
# around any symlinks in the current WD's path
source_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

function ts {
  set +x
  echo "$(date +%H:%M:%S.%N):" "${@}"
  set -x
}
export -f ts

function init_go_env_vars {
  if [[ -z "${GOPATH}" ]]; then
    export GOPATH="${HOME}/go"
  fi

  if [[ -z "${GOBIN}" ]]; then
    export GOBIN="${GOPATH}/bin"
  fi
}

############################
# Behavior tables for args #
############################
# +======================+=====================+==================+
# | Initial State / Flag | (no --restart-kube) |  --restart-kube  |
# |                      |                     |     (default)    |
# +======================+=====================+==================+
# |                      |                     |                  |
# |   No minikube        |  minikube start     |  minikube start  |
# |                      |                     |                  |
# |                      |                     |                  |
# +----------------------+---------------------+------------------|
# |   minikube           |                     |  minikube delete |
# |                      |                     |  minikube start  |
# |                      |                     |                  |
# +----------------------+---------------------+------------------|
#
# # NOTE: Pach options should be monotonic (0-4)
# +======================+================+================+================+================+
# | Initial State / Flag | --no-pachyderm |   PUSH_PACH    | RESTART_PACH   | REDEPLOY_PACH  |
# |                      |                |                |                |   (default)    |
# +======================+================+================+================+================+
# |   No Pachyderm       |                | build & push   | build & push   | build & push   |
# |                      |                |                | helm|kc create | helm|kc create |
# |                      |                |                |                |                |
# +----------------------+----------------+----------------+----------------+----------------+
# |     Pachyderm        | helm|kc delete | build & push   | build & push   | build & push   |
# |                      |                | # restart      | # restart      | helm|kc delete |
# |                      |                | kc delete po   | kc delete po   | helm|kc create |
# +----------------------+----------------+----------------+----------------+----------------+

declare _NO_PACH=0
declare _PUSH_PACH=1
declare _RESTART_PACH=2
declare _REDEPLOY_PACH=3

# Parse script arguments. Sets the global variables:
# - HELM_VALUES
# - PACH_ACTION
# - RESTART_KUBERNETES
# - BUILD_MOUNT_SERVER
# - KUBE_VERSION
# - PACH_VERSION
function parse_args {
  # parse args in variable definition instead of eval so that "$?" is set
  # correctly
  local new_args  # assigning inline would mask "missing argument" errors
  new_args="$(getopt -o "n:" -l "deploy-args:,no-pachyderm,only-pachyderm,tag:,version:,kubernetes-version:,push-only" -- "${@}")"
  # shellcheck disable=SC2181
  if [[ "$?" -ne 0 ]]; then
    exit 1
  fi
  eval "set -- ${new_args}"
  declare -g HELM_VALUES=()  # If unset, don't pass --set. If set, pass --set as extra arg
  declare -g PACH_ACTION="${_REDEPLOY_PACH}"
  declare -g RESTART_KUBERNETES=true
  declare -g BUILD_MOUNT_SERVER=true
  declare -g KUBE_VERSION=""
  declare -g PACH_VERSION=""
  declare -ag K8S_ARGS
  local kube_namespace
  while true; do
    case "${1}" in
      -n)
          # will validate at the end and write to K8S_ARGS
          kube_namespace="${2}"
          shift 2
          ;;
      --no-pachyderm)
          PACH_ACTION="${_NO_PACH}"
          shift
          ;;
      --push-pach-image)
          PACH_ACTION="${_PUSH_PACH}"
          shift
          ;;
      --update-pach-pod)
          PACH_ACTION="${_RESTART_PACH}"
          shift
          ;;
      --no-restart-kube)
          unset RESTART_KUBERNETES
          shift
          ;;
      --no-build-mount-server)
          unset BUILD_MOUNT_SERVER
          shift
          ;;
      --deploy-args)
          HELM_VALUES=( "--set" "${2}" )
          shift 2
          ;;
      --version)
          if ! [[ "${2}" =~ [0-9]+\.[0-9]+\.[0-9] ]]; then
            echo "--tag must be of the form 'A.B.C'" >/dev/stderr
            exit 1
          fi
          if [[ -n "${PACH_VERSION}" ]]; then
            echo "version already set to ${PACH_VERSION}" >/dev/stderr
            exit 1
          fi
          PACH_VERSION="${2}"
          shift 2
          ;;
      --tag)
          if ! [[ "${2}" =~ v.* ]]; then
            echo "--tag must be of the form 'vA.B.C'"
            exit 1
          fi
          if [[ -n "${PACH_VERSION}" ]]; then
            echo "version already set to ${PACH_VERSION}" >/dev/stderr
            exit 1
          fi
          PACH_VERSION="${2##v}"
          shift 2
          ;;
      --kubernetes-version)
        KUBE_VERSION="${2}"
        shift 2
        ;;
      --)
        shift
        break
        ;;
       *)
        exit 1
        ;;
    esac
  done

  if [[ -z "${PACH_VERSION}" ]]; then
    PACH_VERSION=local
    if [[ "${PWD}" != "$(git rev-parse --show-toplevel)" ]] || ! ls ./mascot.txt; then
      echo "Must be in a Pachyderm client"
      exit 1
    fi
  fi

  if [[ -n "${kube_namespace}" ]] && ! kubectl get namespace "${2}"; then
    echo "Could not create resources in k8s namespace \"${kube_namespace}\"; does not exist"
    exit 1
  fi
  local response
  if [[ -z "${kube_namespace}" ]] \
    && kubectl get namespace "test-cluster-1" >/dev/null 2>&1 \
    && [[ "${PACH_VERSION}" == "local" ]] \
    && [[ "${PACH_ACTION}" == "${_REDEPLOY_PACH}" ]] \
    && [[ "${RESTART_KUBERNETES}" == "true" ]]; then
    echo "It looks like you're iterating on pachd tests. Would you like to"
    echo "limit this reset to:"
    echo "  1) clearing the 'test-cluster-1' namespace and"
    echo "  2) pushing a new image?"
    echo
    echo "(y/N)"
    if ! read -rt 10 response; then
      response="n"
    fi
    if [[ "${response:0:1}" == "y" ]] || [[ "${response:0:1}" == "Y" ]]; then
      kube_namespace="test-cluster-1"
      PACH_ACTION="${_PUSH_PACH}"
      unset RESTART_KUBERNETES
    fi
  elif minikube status \
    && kubectl get po -l app=pachd,suite=pachyderm >/dev/null 2>&1 \
    && [[ "${PACH_VERSION}" == "local" ]] \
    && [[ "${PACH_ACTION}" == "${_REDEPLOY_PACH}" ]] \
    && [[ "${RESTART_KUBERNETES}" == "true" ]]; then
    if ! read -rt 10 -p \
      "Would you like to restart only the pachd pod, instead of all of minikube? (y/N) " \
      response; then
      response="n"
    fi
    if [[ "${response:0:1}" == "y" ]] || [[ "${response:0:1}" == "Y" ]]; then
      PACH_ACTION="${_RESTART_PACH}"
      unset RESTART_KUBERNETES
    fi
  fi

  if [[ -n "${kube_namespace}" ]]; then
    K8S_ARGS=( -n "${kube_namespace}" )
  fi

  # These vars are used in xargs subshells and so need to be exported
  export PACH_ACTION
  export PACH_VERSION
  export BUILD_MOUNT_SERVER
}

# print pachyderm's kubernetes manifest (for undeploy/deploy)
#
# Sets the global variable:
# - HELM_TEMPLATE_COMMAND
function set_helm_command {
  declare -ag HELM_TEMPLATE_COMMAND
  # Note: this activates auth by default. You can:
  # - remove --set pachd.enterpriseLicenseKey=..., or
  # - add --set pachd.activateAuth=false
  # to change this
  local chart_source
  if [[ "${PACH_VERSION}" != "local" ]]; then
    # Instead of
    # 'helm repo add pachrepo; helm template pach pachrepo/pachyderm ...'
    # we just use the .tgz url directly. This can be obtained by looking at
    # https://helm.pachyderm.com/index.yaml.
    # See the docs at https://v2.helm.sh/docs/chart_repository/
    # This may run outside a Pachyderm git repo
    chart_source="https://github.com/pachyderm/helmchart/releases/download/pachyderm-${PACH_VERSION}/pachyderm-${PACH_VERSION}.tgz"
  else
    chart_source="./etc/helm/pachyderm"
  fi
  HELM_TEMPLATE_COMMAND=(
    helm template pach "${chart_source}"
      --set deployTarget=LOCAL
      --set pachd.enterpriseLicenseKey="${ENT_ACT_CODE}"
      --set pachd.image.tag=local
      "${HELM_VALUES[@]}"
  )
  export HELM_TEMPLATE_COMMAND
}

function maybe_delete_minikube {
  # Start minikube restart process
  if [[ "${RESTART_KUBERNETES}" == true ]]; then
    minikube delete
  fi
  ts "minikube delete is done"
}

# Function for creating a kind cluster with a local registry intead of
# minikube. Unused for now, as kind seems to be much, much slower than minikube
# for some reason (in particular, pushing images to kind is very slow on my
# machine. It's faster with the local registry but still much slower than
# minikube)
function maybe_create_kind_cluster_with_local_registry {
  # Code below copied from https://kind.sigs.k8s.io/docs/user/local-registry/
  # create registry container unless it already exists
  local reg_name='kind-registry'
  local reg_port='5001'
  if ! docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null; then
    docker run \
      -d --restart=always -p "127.0.0.1:${reg_port}:5000" --name "${reg_name}" \
      registry:2
  fi

  ## create a cluster with the local registry enabled in containerd
  # the guide above makes this a mirror of the 'localhost:5001' registry, but
  # I want images to be pulled from my local registry by default (breaking all
  # kinds of shit but allowing the pachyderm helm chart to do the right thing,
  # since it doesn't prefix the images within with 'localhost:5001') so I
  # change the containerd config to make my registry a mirror of docker.io
  #
  # Per some old CRI docs (https://github.com/containerd/cri/blob/8f1a8a1fb9ebd821a1afe3b3ff3adec7bd33cfdf/docs/registry.md):
  # "The endpoint is a list that can contain multiple image registry URLs split
  # by commas. When pulling an image from a registry, containerd will try these
  # endpoint URLs one by one, and use the first working one."

  #   cat <<EOF | kind create cluster --name=pachd-client-cluster --config=-
  # kind: Cluster
  # apiVersion: kind.x-k8s.io/v1alpha4
  # containerdConfigPatches:
  # - |-
  #   [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${reg_port}"]
  #     endpoint = ["http://${reg_name}:5000"]
  cat <<EOF | kind create cluster --name=pachd-client-cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
    endpoint = ["http://${reg_name}:5000", "https://index.docker.io"]
EOF

  # connect the registry to the cluster network if not already connected
  if [ "$(docker inspect -f='{{json .NetworkSettings.Networks.kind}}' "${reg_name}")" = 'null' ]; then
    docker network connect "kind" "${reg_name}"
  fi

  # Document the local registry
  # https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/generic/1755-communicating-a-local-registry
  cat <<EOF | kubectl "${K8S_ARGS[@]}" apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${reg_port}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF
}

function maybe_start_minikube {
  if minikube status; then
    return
  fi
  if [[ -n "${KUBE_VERSION}" ]]; then
    minikube start --disk-size=15g --kubernetes-version="${KUBE_VERSION}"
  else
    minikube start --disk-size=15g
  fi
}

# get_old_pach_version gets the version of pach that's running from the
# existing pachd pod, if any.
#
# Sets the global variable:
# - OLD_PACHD_POD
# - OLD_PACH_VERSION.
function get_old_pach_version {
  declare -g OLD_PACHD_POD
  declare -g OLD_PACH_VERSION
  if kubectl version; then
    # See if we just need to restart the pachd pod, or if we need to undeploy+redeploy
    OLD_PACHD_POD="$( pachd_pod "${K8S_ARGS[@]}" )"
    if [[ "${OLD_PACHD_POD}" = "NotFound" ]]; then
      unset OLD_PACHD_POD
    fi
    if [[ -n "${OLD_PACHD_POD}" ]]; then
      OLD_PACH_VERSION="$(kubectl "${K8S_ARGS[@]}" get deploy/pachd -o json | jq -r '.spec.template.spec.containers[0].image' | sed 's#^pachyderm/pachd:##' )"
    fi
  fi
}

function maybe_undeploy_pachyderm {
  if [[ "${PACH_ACTION}" != "${_NO_PACH}" ]] && [[ "${PACH_ACTION}" != "${_REDEPLOY_PACH}" ]]; then
    return
  fi
  if [[ -n "${OLD_PACHD_POD}" ]]; then
    "${HELM_TEMPLATE_COMMAND[@]}" | kubectl "${K8S_ARGS[@]}" delete --ignore-not-found=true -f -
  fi
}

function maybe_build_or_pull_pachyderm {
  if [[ "${PACH_ACTION}" -lt "${_RESTART_PACH}" ]]; then
    return
  fi

  ## Common case: building pach locally for testing
  # build pachctl/pachd in parallel
  if [[ "${PACH_VERSION}" == "local" ]]; then
    declare -a build_items
    build_items=( "make install" "make docker-build-amd" )
    if [[ "${BUILD_MOUNT_SERVER}" == "true" ]]; then
      # Soon...
      # build_items+=("make mount-server")
      # shellcheck disable=SC2016
      build_items+=('CGO_ENABLED=0 go install -gcflags "all=-trimpath=${PWD}" ./src/server/cmd/mount-server')
    fi
    IFS=$'\n' eval 'echo -n "${build_items[*]}"' \
      | xargs -t -n1 -P0 -d'\n' /bin/bash -c
    return
  fi

  ## Rare case: deploying released version
  ts "downloading pachctl v${PACH_VERSION}"
  if ! which pachctl || [[ "$(pachctl version --client-only)" != "${PACH_VERSION}" ]]; then
    # Download the appropriate version of pachctl
    curl -L https://github.com/pachyderm/pachyderm/releases/download/v${PACH_VERSION}/pachctl_${PACH_VERSION}_linux_amd64.tar.gz \
      | tar -C "${GOBIN}" --strip-components=1 -xzf - pachctl_${PACH_VERSION}_linux_amd64/pachctl
  fi

  # docker pull pachd and worker
  for i in pachd worker; do
    docker pull pachyderm/${i}:${PACH_VERSION}
  done
  return

}

function maybe_push_images_to_kube {
  if [[ "${RESTART_KUBERNETES}" == true ]]; then
    # Minikube is new, so we must re-pull & deploy accessory images
    for image in \
      "busybox:1.28" \
      "postgres:13.0-alpine" \
      "edoburu/pgbouncer:1.15.0" \
      "docker.io/bitnami/postgresql:13.3.0" \
      "quay.io/coreos/etcd:v3.3.5" \
      "jaegertracing/all-in-one:1.10.1"
    do
  		{ docker images | kmp "${image}"; } || docker pull "${image}"
      "${source_dir}/../kube/push-to-minikube.sh" "${image}"
    done
  fi

  if [[ "${DEPLOY_PACHYDERM}" = "true" ]]; then
    "${source_dir}/../kube/push-to-minikube.sh" pachyderm/pachd:${PACH_VERSION}
    "${source_dir}/../kube/push-to-minikube.sh" pachyderm/worker:${PACH_VERSION}
  fi
}

function maybe_deploy_pachyderm {
  if [[ "${PACH_ACTION}" -lt "${_RESTART_PACH}" ]]; then
    return
  fi

  # short circuit - if we're just restarting 'pachd:local', we can restart the
  # pod so it uses the new image (which should've been pushed already)
  # Note: if ${OLD_PACH_VERSION != "local", then we have to fully undeploy, as
  # restarting that pod won't do anything.
  if [[ "${PACH_ACTION}" == "${_RESTART_PACH}" ]] \
    && [[ -n "${OLD_PACHD_POD}" ]] \
    && [[ "${OLD_PACH_VERSION}" == "local" ]]; then
    ts "Restarting only the pachd pod (not minikube)"
    # Just delete the old pod--pushing a new image with the same tag will cause
    # the pod to be recreated using the new image
    kubectl "${K8S_ARGS[@]}" delete po -l app=pachd,suite=pachyderm --ignore-not-found=true # restart pachd

    # Fully wait for old pod to die and new pod to accept requests, otherwise
    # tests connect to the old pod and then crash when it closes its connections
    local state
    ts "Waiting for old pachd pod to die..."
    for i in $(seq 100); do
      state="$(kubectl "${K8S_ARGS[@]}" get po/"${OLD_PACHD_POD}" -o 'jsonpath={.items[].status.phase}' 2>&1 || true)"
      if [[ "${state}" =~ "NotFound" ]]; then
        return
      fi
      echo "Waiting for old pod (still in ${state}) (try $(( i+1 ))/100)"
    done
    ts "Old pachd pod (${OLD_PACHD_POD}) never died; continuing..."
    return
  fi

  # Deploy new Pachyderm
  "${HELM_TEMPLATE_COMMAND[@]}" | kubectl "${K8S_ARGS[@]}" apply -f -
}

function maybe_connect_to_pachyderm {
  if [[ "${PACH_ACTION}" -lt "${_REDEPLOY_PACH}" ]]; then
    return
  fi

  # Clear pach config
  echo "Deleting pach config, to connect to new cluster"
  [[ -z "${PACH_CONFIG}" ]] && PACH_CONFIG="${HOME}/.pachyderm/config.json"
  if [[ -f "${PACH_CONFIG}" ]]; then
    cp "${PACH_CONFIG}" "${PACH_CONFIG}.bak" || true
    rm "${PACH_CONFIG}"
  fi

  # Extract Pach root token from the deployment above and use it in the local
  # pachctl
  if kubectl "${K8S_ARGS[@]}" get secret/pachyderm-auth; then
    # Add trailing newline, so that 'pachctl auth' is able to read the token
    # shellcheck disable=SC1003
    kubectl "${K8S_ARGS[@]}" get secret/pachyderm-auth -o jsonpath='{.data.root-token}' \
      | base64 -d \
      | sed -e '$a\' \
      | pachctl auth use-auth-token
  else
    # At some point, the name of the auth token that Pachyderm deploys changed.
    # I don't know whether this or the above is newer or what
    # shellcheck disable=SC1003
    kubectl "${K8S_ARGS[@]}" get secret/pachyderm-bootstrap-config -o jsonpath='{.data.rootToken}' \
      | base64 -d \
      | sed -e '$a\' \
      | pachctl auth use-auth-token
  fi

  # Turn pachyderm proxy service into a NodePort
  # (minikube tunnel doesn't work on Linux for some reason)
  # This is preferred over the built-in port-forward mechanism because it makes
  # testing jupyter easier
  if kubectl "${K8S_ARGS[@]}" get svc/pachyderm-proxy; then
    kubectl get svc/pachyderm-proxy -o json \
      | jq '.spec.type="NodePort"' \
      | kubectl apply -f -
    pachctl config update context \
      --pachd-address="$(minikube ip):$(kubectl get svc/pachyderm-proxy -o jsonpath='{$.spec.ports[0].nodePort}')"
  fi

  # Pachd RC already exists & pod should come back on its own
  # Wait for pachd pod to enter "Running"
  local new_pod
  while true; do
    sleep 1
    new_pod="$( pachd_pod "${K8S_ARGS[@]}" )"
    if [[ "${new_pod}" == "NotFound" ]]; then
      continue
    fi
    state="$(kubectl "${K8S_ARGS[@]}" get po/"${new_pod}" -o 'jsonpath={.status.phase}' 2>&1 || true)"
    if [[ "${state}" == "Running" ]]; then
      break
    fi
    ts "Waiting for new pod (still in ${state})"
  done
  ts "pachd/worker pods are in RUNNING"

  # Wait for pachctl to connect
  until pachctl version; do
    echo "No Pachyderm at ${PACHD_ADDRESS}"
    sleep 1
  done
  ts "pachd is available, logging in..."
}

function __main__ {
  init_go_env_vars
  parse_args "$@"
  set_helm_command # depends on args (e.g. version="local")

  ts "Starting"
  # Now that validation is done, start showing all commands
  set -x
  maybe_delete_minikube

  # Start minikube and obtain pachd/pachctl in parallel
  export -f maybe_start_minikube
  export -f maybe_build_or_pull_pachyderm
  echo -e "maybe_start_minikube\nmaybe_build_or_pull_pachyderm" \
    | xargs -t -n1 -P0 -d'\n' /bin/bash -c
  maybe_push_images_to_kube
  ts "pachd/worker pushed"

  ###
  # Wait for minikube to come up and for pachctl (and the pachd/worker images) to
  # finish building (the prereqs for deploying)
  ###
  set +x
  WHEEL='\-/|'; W=0
  until minikube status; do
    echo -en "\e[G${WHEEL:$((W=(W+1)%4)):1} Waiting for Minikube to come up..."
    sleep 1
  done
  ts "minikube is available"

  until pachctl version --client-only >/dev/null 2>&1; do
    echo -en "\e[G${WHEEL:$((W=(W+1)%4)):1} Waiting for pachctl to build..."
    hash -r
    sleep 1
  done
  hash -r
  ts "pachctl is built"

  echo "###"
  echo "# Deploying pachyderm version v$(pachctl version --client-only)"
  echo "###"
  set -x

  # Deploy minio (what we use for local deployments now). This is necessary
  # even when pachyderm isn't being deployed (and goes into the default
  # namespace, so no $K8S_ARGS) as minikubetestenv, which we use for unit tests
  # and obviates the need to deploy pachyderm directly, expects it.
  kubectl apply -f "$(git root)/etc/testing/minio.yaml"

  ###
  # Deploy pachyderm into minikube if needed
  ###
  get_old_pach_version
  maybe_undeploy_pachyderm
  maybe_deploy_pachyderm
  maybe_connect_to_pachyderm

  ts "minikube is up"
  notify-send -i /home/mjs/Pachyderm/logo_little.png "Minikube is up" -t 10000
}
__main__ "$@"
