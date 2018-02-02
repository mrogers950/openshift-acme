#!/bin/bash
set -e
shopt -s expand_aliases

wd=$(pwd)

script_full_path=$(readlink -f $0)
script_dir=$(dirname ${script_full_path})
pushd ${script_dir}/..

if [[ "$1" = /* ]]
then
   # Absolute path
   bindir=$1
else
   # Relative path
   bindir=${wd}/${1:-.}
fi
PATH=${bindir}:${PATH}
prefix=oc-
pathPrefix=${bindir}/${prefix}
binaries=$(find ${bindir} -type f -executable -name "${prefix}*" -printf "%T+\t%p\n" | sort | cut -d$'\t' -f2)
echo binaries: ${binaries}

[[ ! -z "${binaries}" ]]

make -j64 test-extended-install GOFLAGS='-v -race'

function setupClusterWide() {
    sed '/^spec:/a \ \ paused: true' -i deploy/letsencrypt-staging/cluster-wide/deployment.yaml
    oc create -fdeploy/letsencrypt-staging/cluster-wide/{clusterrole,serviceaccount,deployment}.yaml
    oc adm policy add-cluster-role-to-user openshift-acme -z openshift-acme
    export FIXED_NAMESPACE=""
}

function setupSingleNamespace() {
    oc login -u developer -p developer
    sed '/^spec:/a \ \ paused: true' -i deploy/letsencrypt-staging/single-namespace/deployment.yaml
    oc create -fdeploy/letsencrypt-staging/single-namespace/{role,serviceaccount,deployment}.yaml
    oc policy add-role-to-user openshift-acme --role-namespace="$(oc project --short)" -z openshift-acme
    export FIXED_NAMESPACE=$(oc project --short)
    oc login -u system:admin

    # hack; needs bug fix https://github.com/openshift/origin/pull/18312
    # adds update permissions for custom-host
    oc patch role openshift-acme --type=json -p='[{"op": "add", "path": "/rules/1/verbs/1", "value": "update"}]'
}

function failureTrap() {
    oc get nodes
    oc get all -n default
    oc get all
    oc describe deploy/openshift-acme
    oc get routes,svc --all-namespaces
    oc get events
    docker images
    oc get po -o yaml
    oc logs deploy/openshift-acme || true

    docker logs origin

    sleep 3
}

trap failureTrap ERR
trap "sleep 3" EXIT

for binary in ${binaries}; do
    docker rm -f $(docker ps -aq) || true

    version=${binary#$pathPrefix}
    echo binary version: ${version}
    ln -sfn ${binary} ${bindir}/oc
    oc version || true
    for setup in {setupClusterWide,setupSingleNamespace}; do
        echo ${setup}
        oc cluster up --version=${version} --server-loglevel=4
        oc version
        oc login -u system:admin

        oc get all -n default

        # Wait for docker-registry
        # Wait for router
        (timeout 5m bash -c 'oc rollout status -n default dc/docker-registry && oc rollout status -n default dc/router') || (\
        oc get all -n default; \
        oc get -n default po/docker-registry-1-deploy po/router-1-deploy -o yaml; \
        oc get nodes; \
        docker logs origin; \
        sleep 3 \
        false)

        oc new-project acme-aaa
        oc get sa,secret

        # Create ImageStream from the image build earlier
        sa_secret_name=$(oc get sa builder --template='{{ (index .imagePullSecrets 0).name }}')
        token=$(oc get secret ${sa_secret_name} --template='{{index .metadata.annotations "openshift.io/token-secret.value"}}')
        registry=$(oc get svc/docker-registry -n default --template='{{.spec.clusterIP}}:{{(index .spec.ports 0).port}}')
        docker login -u aaa -p ${token} ${registry}
        is_image=${registry}/$(oc project --short)/openshift-acme
        docker tag openshift-acme-candidate ${is_image}
        docker push ${is_image}

        oc get is openshift-acme -o yaml

        ${setup}
        oc set env -e OPENSHIFT_ACME_DEFAULT_ROUTE_TERMINATION=Allow deploy/openshift-acme

        oc rollout resume deploy/openshift-acme
        sleep 10
        oc get deploy/openshift-acme -o yaml
        oc rollout status deploy/openshift-acme
        oc get all

        make -j64 test-extended GOFLAGS="-v -race" GO_ET_KUBECONFIG=~/.kube/config GO_ET_DOMAIN=${DOMAIN} || (oc logs deploy/openshift-acme; false)
        oc get all
        oc logs deploy/openshift-acme

        oc get deploy/openshift-acme --template='deployed: {{(index .spec.template.spec.containers 0).image}}'
        docker images

        oc cluster down
    done
done
