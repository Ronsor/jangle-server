#!/bin/sh

grep -E '\&APIResponseError\{[0-9]{5}' *.go
