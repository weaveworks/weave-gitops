#!/bin/bash

# This is a deliberately opinionated script for developing Weave Gitops.
# To get started run ./tools/reboot.sh --help
#
# WARN: This script is designed to be "turn it off and on again". It will delete
# the given kind cluster (if it exists) and recreate, installing everything from
# scratch.

export KIND_CLUSTER_NAME=${KIND_CLUSTER_NAME:-wego-dev}
NAMESPACE=flux-system

do_kind() {
        kind delete cluster --name "$KIND_CLUSTER_NAME"

        "$(dirname "$0")/kind-with-registry.sh" || exit 1
}

do_bootstrap(){
        local provider="$1"
        local owner="$2"
        local repo="$3"

        flux bootstrap "$provider" \
            --owner="$owner" \
            --repository="$repo" \
            --branch=main \
            --path=./clusters/wego-dev \
            --personal
}

do_install(){
        local namespace="$1"

        flux install -n "$namespace"
}

must_set() {
        local var="$1"
        local name="$2"
        local cmd="$3"

        if [[ -z ${var} ]]; then
            printf "Must set %s\n\n" "$name"

            cmd_"$cmd"_help
            exit 1
        fi
}

cmd_install() {
        local namespace="$NAMESPACE"

        while [ $# -gt 0 ]; do
                case "$1" in
                "-h" | "--help")
                        cmd_install_help
                        exit 0
                        ;;
                "-n" | "--namespace")
                        shift
                        namespace="$1"
                        ;;
                *)
                        echo "Unknown argument: $1. Please use --help for help."
                        exit 1
                        ;;
                esac
                shift
        done

        echo "Will destroy and recreate cluster $KIND_CLUSTER_NAME and install flux to namespace $namespace."

        do_kind "$@"
        do_install "$namespace"
}

cmd_bootstrap() {
        local provider="github"
        local owner=""
        local repo=""

        while [ $# -gt 0 ]; do
                case "$1" in
                "-h" | "--help")
                        cmd_bootstrap_help
                        exit 0
                        ;;
                -p | --git-provider)
                        shift
                        provider="$1"
                        ;;
                -o | --owner)
                        shift
                        owner="$1"
                        ;;
                -r | --repo)
                        shift
                        repo="$1"
                        ;;
                *)
                        echo "Unknown argument: $1. Please use --help for help."
                        exit 1
                        ;;
                esac
                shift
        done

        must_set "$owner" "owner" "bootstrap"
        must_set "$repo" "repo" "bootstrap"
        must_set "$GITHUB_TOKEN" "GITHUB_TOKEN" "bootstrap"

        echo "Will destroy and recreate cluster $KIND_CLUSTER_NAME and bootstrap flux with gitops repo $provider.com/$owner/$repo."

        do_kind "$@"
        do_bootstrap "$provider" "$owner" "$repo"
}

cmd_install_help() {
        cat <<EOF
  install                  Install flux to the kind cluster. This is the default action. If you run \`reboot\` without any commands, install will just run.
    OPTIONS:
      --namespace, -n      Namespace to install flux to. default: $NAMESPACE

EOF
}

cmd_bootstrap_help() {
        cat <<EOF
  bootstrap                Bootstrap flux in the kind cluster and connect to a given gitops repo
    OPTIONS:
      --git-provider, -p   Git provider. default: github
      --owner, -o          Owner of gitops repo
      --repo, -r           Name of gitops repo

EOF
}

cmd_help() {
        cat <<EOF
Usage: $0 <COMMAND> <OPTIONS>

Script to tear down and reprovision a dev environment for weave-gitops.

WARN: will destroy and recreate the kind cluster.

Note: this script is deliberately bare-bones. It will not be extended.

ENV VARS:
  GITHUB_TOKEN              With repo scope. Required for \`bootstrap\` command.
  KIND_CLUSTER_NAME         Set to configure created kind cluster name. default: wego-dev

COMMANDS:
EOF

        cmd_install_help
        cmd_bootstrap_help
}

main() {
        if [ $# = 0 ]; then
            echo "No command provided. Will destroy and recreate cluster $KIND_CLUSTER_NAME and install flux to namespace $NAMESPACE."
        fi

        local default="${1:-install}"

        while [ $# -gt 0 ]; do
                case "$1" in
                -h | --help)
                        cmd_help
                        exit 0
                        ;;
                -*)
                        echo "Unknown arg: $1. Please use \`$0 help\` for help."
                        exit 1
                        ;;
                *)
                        break
                        ;;
                esac
                shift
        done

        cmd=cmd_$default

        # Check if given cmd is valid
        # shellcheck disable=SC2091
        if ! $(declare -f "$cmd" >/dev/null) ; then
            echo "Unknown command: $1. Please use \`$0 help\` for help." && exit 1
        fi

        # make $@ a list of command-specific args
        shift

        # Run it
        if $cmd "$@" ; then
            echo "Done."
            echo "You can now run \`make cluster-dev\` or install the Weave Gitops Chart or whatever."
        fi
}

main "$@"
