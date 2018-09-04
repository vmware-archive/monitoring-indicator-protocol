#! /bin/bash

go clean -cache
go test ./... -v

exit_status=$?

if [ $exit_status -ne 0 ]; then
    echo "TESTS FAILED!"
fi

exit $exit_status
