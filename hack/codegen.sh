#!/bin/sh
bash ./vendor/k8s.io/code-generator/generate-groups.sh all github.com/s1061123/multus-proxy/pkg/client github.com/s1061123/multus-proxy/pkg/apis multus:v1alpha --output-package="k8s.io/client-go"
