#!/usr/bin/env bash
#
# gen-test-target will generate a golang testcase using the
# kustomize test-harness that is used from kerbernetes-sigs/pkg/kusttest/kusttestharness.go
# The unittest compares the collection of resource files with what kustomize build would produce (actual vs expected)
#
source hack/utils.sh

kebab-case-2-PascalCase() {
  local a=$1 b='' array
  IFS='-' read -r -a array <<< "$a"
  for element in "${array[@]}"; do
    part="${element}"
    part="$(tr '[:lower:]' '[:upper:]' <<< ${part:0:1})${part:1}"
    b+=$part
  done
  echo $b
}

gen-target-start() {
  local dir=$(get-target $1) target fname
  fname=/manifests${dir##*/manifests}
  target=$(kebab-case-2-PascalCase $(get-target-name $1))

  echo 'package tests_test'
  echo ''
  echo 'import ('
  echo '  "sigs.k8s.io/kustomize/k8sdeps/kunstruct"'
  echo '  "sigs.k8s.io/kustomize/k8sdeps/transformer"'
  echo '  "sigs.k8s.io/kustomize/pkg/fs"'
  echo '  "sigs.k8s.io/kustomize/pkg/loader"'
  echo '  "sigs.k8s.io/kustomize/pkg/resmap"'
  echo '  "sigs.k8s.io/kustomize/pkg/resource"'
  echo '  "sigs.k8s.io/kustomize/pkg/target"'
  echo '  "testing"'
  echo ')'
  echo ''
  echo 'func write'$target'(th *KustTestHarness) {'
}

gen-target-middle() {
  local directory=$1

  for i in $(echo $(cat $directory/kustomization.yaml | grep '^- .*yaml$' | sed 's/^- //') $(cat $directory/kustomization.yaml | grep '  path: ' | sed 's/^.*: \(.*\)$/\1/') params.env kustomization.yaml | sed 's/ /\\n/g' | sort | uniq | awk '{gsub(/\\n/,"\n")}1'); do
    file=$i
    if [[ -f $directory/$file ]]; then
      case $file in
        "kustomization.yaml")
          gen-target-kustomization $file $directory
          ;;
        *.yaml)
          gen-target-resource $file $directory
          ;;
        params.env)
          gen-target-resource $file $directory
          ;;
        *) ;;

      esac
    fi
  done
}

gen-target-end() {
  echo '}'
}

gen-target() {
  local directory=$1
  gen-target-start $directory
  gen-target-middle $directory
  gen-target-end
}

gen-target-base() {
  echo '  th.writeK("'$kname'", `'
  cat $dir/$file | sed 'sx- ../../basex- '$basedir'x'
  echo '`)'
}

gen-target-kustomization() {
  local file=$1 dir=$2 fname kname basedir
  fname=/manifests${dir##*/manifests}
  kname=${fname%/kustomization.yaml}
  echo '  th.writeK("'$kname'", `'
  cat $dir/$file 
  echo '`)'
  if [[ $(get-target-dirname $dir) != "base" ]]; then
    basedir=$(get-target $dir)/base
    if [[ -f $basedir/kustomization.yaml ]]; then
      gen-target-middle $basedir
    fi
  fi
}

gen-target-resource() {
  local file=$1 dir=$2 fname
  fname=/manifests${dir##*/manifests}/$file

  echo '  th.writeF("'$fname'", `'
  cat $dir/$file | sed 's/`/U+0060/g'
  echo '`)'
}

gen-test-case() {
  local base=$(get-target-name $1) dir=$(get-target $1) target fname
  fname=/manifests${dir##*/manifests}/$(get-target-dirname $1)
  target=$(kebab-case-2-PascalCase $base)
  targetPath=..${1#*/manifests}

  gen-target $1
  echo ''
  echo 'func Test'$target'(t *testing.T) {'
  echo '  th := NewKustTestHarness(t, "'$fname'")'
  echo '  write'$target'(th)'
  echo '  m, err := th.makeKustTarget().MakeCustomizedResMap()'
  echo '  if err != nil {'
  echo '    t.Fatalf("Err: %v", err)'
  echo '  }'
  echo '  targetPath := "'$targetPath'"'

  echo '  fsys := fs.MakeRealFS()'
  echo '    _loader, loaderErr := loader.NewLoader(targetPath, fsys)'
  echo '    if loaderErr != nil {'
  echo '      t.Fatalf("could not load kustomize loader: %v", loaderErr)'
  echo '    }'
  echo '    rf := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()))'
  echo '    kt, err := target.NewKustTarget(_loader, rf, transformer.NewFactoryImpl())'
  echo '    if err != nil {'
  echo '      th.t.Fatalf("Unexpected construction error %v", err)'
  echo '    }'
  echo '  n, err := kt.MakeCustomizedResMap()'
  echo '  if err != nil {'
  echo '    t.Fatalf("Err: %v", err)'
  echo '  }'
  echo '  expected, err := n.EncodeAsYaml()'
  echo '  th.assertActualEqualsExpected(m, string(expected))'
  echo '}'
}

gen-test-case $1
