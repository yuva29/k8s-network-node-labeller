#!/bin/bash

if [ -z $RELEASE ]
then
  echo "RELEASE is not set, return"

  if [ -z ${DOCKERHUB_TOKEN-} ]
  then
      echo "DOCKERHUB_TOKEN is not set"
  else
      echo "DOCKERHUB_TOKEN is set"
  fi

  exit 0
fi

tag_prefix="${RELEASE%-*}"

if [ "$tag_prefix" == "main" ]; then
  tag="latest"
else
  tag="$tag_prefix"
fi

echo "Copying network-node-labeller artifacts..."

setup_dir () {
    ls -al /k8s-network-node-labeller/
    BUNDLE_DIR=/k8s-network-node-labeller/output/
    mkdir -p $BUNDLE_DIR
}

copy_artifacts () {
    # copy node-labeller binary
    cp /k8s-network-node-labeller/build/network-node-labeller $BUNDLE_DIR/k8s-network-node-labeller-$RELEASE.gobin
    # copy docker image
    cp /k8s-network-node-labeller/build/docker/k8s-network-node-labeller-latest.tar.gz $BUNDLE_DIR/k8s-network-node-labeller-$RELEASE.tar.gz
    # copy k8s helm packages
    cp /k8s-network-node-labeller/helm-charts-k8s/internal-k8s-network-node-labeller-helm-k8s-v0.0.1.tgz $BUNDLE_DIR/internal-k8s-network-node-labeller-helm-k8s-v0.0.1-$RELEASE.tgz
    cp /k8s-network-node-labeller/helm-charts-k8s/amdpsdo-k8s-network-node-labeller-helm-k8s-v0.0.1.tgz $BUNDLE_DIR/amdpsdo-k8s-network-node-labeller-helm-k8s-v0.0.1-$RELEASE.tgz
    if [ "$?" -eq "0" ]; then
      echo "Network-NL image copy success"
    else
      echo "Network-NL image copy failed"
      exit $?
    fi
    ls -la $BUNDLE_DIR
}

docker_push () {
    NETWORK_NODE_LABELLER_IMAGE_URL=registry.test.pensando.io:5000/k8s-network-node-labeller

    docker load -i /k8s-network-node-labeller/build/docker/k8s-network-node-labeller-latest.tar.gz
    docker inspect $NETWORK_NODE_LABELLER_IMAGE_URL:latest | grep "HOURLY"
    docker tag $NETWORK_NODE_LABELLER_IMAGE_URL:latest $NETWORK_NODE_LABELLER_IMAGE_URL:$tag
    docker push $NETWORK_NODE_LABELLER_IMAGE_URL:$tag

    if [ -z $DOCKERHUB_TOKEN ]
    then
      echo "DOCKERHUB_TOKEN is not set"
    else
      # rhel 9.4
      docker login --username=shreyajmeraamd --password-stdin <<< $DOCKERHUB_TOKEN
      docker tag $NETWORK_NODE_LABELLER_IMAGE_URL:$tag amdpsdo/k8s-network-node-labeller:$RELEASE
      docker push amdpsdo/k8s-network-node-labeller:$RELEASE
    fi
}

helm_push () {
    if [ -z $DOCKERHUB_TOKEN ]
    then
      echo "DOCKERHUB_TOKEN is not set"
    else
      helm registry login docker.io --username=shreyajmeraamd --password-stdin <<< $DOCKERHUB_TOKEN
      helm push /k8s-network-node-labeller/helm-charts-k8s/amdpsdo-k8s-network-node-labeller-helm-k8s-v0.0.1.tgz oci://docker.io/amdpsdo
    fi
}

setup () {
    setup_dir
    copy_artifacts
    docker_push
    helm_push
}

upload () {
    cd $BUNDLE_DIR
    find . -type f -print0 | while IFS= read -r -d $'\0' file;
      do asset-push builds hourly-k8s-network-node-labeller $RELEASE "$file" ;
      if [ $? -ne 0 ]; then
        exit 1
      fi
    done
}

main () {
  setup
  upload
}

main
exit 0
