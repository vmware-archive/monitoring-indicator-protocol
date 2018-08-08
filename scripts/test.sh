#! /bin/bash

go clean -cache
vgo test ./... -v
