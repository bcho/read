#!/bin/bash

GLIDE=$GOPATH/bin/glide

go get github.com/Masterminds/glide
$GLIDE up
