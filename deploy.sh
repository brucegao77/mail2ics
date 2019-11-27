#!/bin/bash
# when project has modified

# image
IMG="mail2ics"

# container
CONT="mail2ics"

echo "Copying record.json >>>>>>>>>>>>>>>>>>>>>>>>>>>"
docker cp ${CONT}:/root/mail2ics/record.json /root/mail2ics/

echo "Updating project >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>"
git pull

echo "Build & Deploy   >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>"
cd ~/mail2ics || exit
rm mail2ics
echo "Building..."
CGO_ENABLED=0 go build
echo "Deploying..."
docker container kill ${CONT} || true
docker container rm ${CONT} || true
docker image rm ${IMG} || true
docker image build -t ${IMG} .
docker container run --network=host -d --name ${CONT} ${IMG}

exit
