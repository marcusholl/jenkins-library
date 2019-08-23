#!/bin/bash

rm -rf documentation/docs-*
documentation/bin/createDocu.sh $*
docker run --rm -it -v `pwd`:/docs -w /docs/documentation squidfunk/mkdocs-material:3.0.4 build --clean --verbose --strict

