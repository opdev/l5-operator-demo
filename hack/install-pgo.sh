#!/bin/bash

sudo git clone https://github.com/CrunchyData/postgres-operator-examples.git

kubectl apply -k postgres-operator-examples/kustomize/install/namespace