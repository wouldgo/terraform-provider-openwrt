#!/usr/bin/env bash

mkdir --parents ../examples/

if command -v terraform >/dev/null 2>&1; then
  terraform fmt -recursive ../examples/
else
  tofu fmt -recursive ../examples/
fi
