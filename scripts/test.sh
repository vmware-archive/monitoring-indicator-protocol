#!/bin/bash

trap "echo Exited!; exit 1;" SIGINT SIGTERM

PROJECT_DIR="$(cd "$(dirname "$0")/.."; pwd)"
PKG="github.com/pivotal/monitoring-indicator-protocol"

function print_usage {
    echo "usage: test.sh [subcommand] [go test args]"
    echo
    echo -e "\033[1mSubcommands:\033[0m"
    echo "   unit          Run the unit tests"
    echo "   integration   Run the integration tests"
    echo "   local         Run the local tests, that is, integration and unit tests"
    echo "   e2e           Run the end-to-end tests"
    echo "   k8s_e2e       Run the k8s end-to-end tests"
    echo "   bosh_e2e      Run the bosh end-to-end tests"
    echo "   build         Build all binaries for the project"
    echo "   cleaners      Run tools that clean the code base"
}

function print_checkpoint {
    echo
    bold_blue "==================================  $@"
}

function green {
    echo -e "\033[32m$1\033[0m"
}

function red {
    echo -e "\033[31m$1\033[0m"
}

function bold_blue {
    echo -e "\033[1;34m$1\033[0m"
}

function check_output {
    eval "$@"
    local status=$?
    exit_on_failure $status
}

function exit_on_failure {
    if [[ $1 -ne 0 ]]; then
        red "SUITE FAILURE"
        exit $1
    fi
}

function exit_on_dirty_git_directory {
    git diff --quiet
    result=$?
    if [[ $1 -ne 0 ]]; then
        red "GIT DIRECTORY IS NOT CLEAN"
        exit $1
    fi
}

function run_build {
    print_checkpoint "Make Sure Packages Compile"
    go test -run xxxxxxxxxxxxxxxxx "$PKG/..." > /dev/null
    exit_code=$?
    if [ $exit_code -ne 0 ]; then
        return $exit_code
    fi
}

function run_unit {
    print_checkpoint "Running Unit Tests"
    go test -mod=vendor -race $(go list ./... | grep -v e2e | grep -v smoke | grep -v cmd)
    exit_code=$?
    return $exit_code
}

function run_integration {
    print_checkpoint "Running Integration Tests"
    go test -mod=vendor -race $(go list ./... | grep cmd)
    exit_code=$?
    return $exit_code
}

function run_local {
    print_checkpoint "Running Local (Unit + Integration) Tests"
    go test -mod=vendor -race $(go list ./...  | grep -v e2e | grep -v smoke)
    exit_code=$?
    return $exit_code
}

function run_k8s_e2e {


    print_checkpoint "Running K8S End-To-End Tests"
    go test -mod=vendor -race "$PKG/k8s/test/e2e" "$@"
    return $?
}

function run_bosh_e2e {
    print_checkpoint "Running Bosh End-To-End Tests"
    go test -mod=vendor -race "$PKG/test/e2e" $@
    return $?
}

function run_e2e {
    run_bosh_e2e
    e2e_exit_code=$?
    run_k8s_e2e
    k8s_exit_code=$?
    return $[ e2e_exit_code + k8s_exit_code ]
}

function parse_argc {
    command=run_local
    if [[ $# -eq 0 ]]; then
        return
    fi

    arg=$1
    case "$arg" in
        -h|-help|--help|help)
            print_usage
            exit 0
            ;;
        unit|integration|local|e2e|bosh_e2e|k8s_e2e|build|cleaners)
            command=run_$arg
            ;;
        *)
            echo "Invalid command: $arg\n"
            print_usage
            exit 1
            ;;
    esac
}

function clear_cache {
    # https://github.com/golang/go/issues/28680
    GO111MODULE=off go clean -cache
}

function main {
    clear_cache
    go mod tidy
    go mod vendor

    parse_argc $1
    shift
    "$command" $@
    result=$?
    exit_on_failure $result
    exit_on_dirty_git_directory $result
    green "SWEET SUITE SUCCESS"
}

main $@
