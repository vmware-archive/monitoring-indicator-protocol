#! /bin/bash

GO111MODULE=off # https://github.com/golang/go/issues/28680
go clean -cache

GO111MODULE=on
go test -mod=vendor ./... -v

exit_status=$?

if [ $exit_status -ne 0 ]; then
    echo "TESTS FAILED!"
fi

exit $exit_status
