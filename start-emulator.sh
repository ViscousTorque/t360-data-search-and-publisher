#!/bin/sh

CONTAINER_NAME=pubsub-emulator
PROJECT_ID=my-gcp-project
TOPIC_ID=my-pubsub-topic
PORT=8085

echo "Starting Pub/Sub Emulator container..."

docker run --rm -d \
  --name $CONTAINER_NAME \
  -e PUBSUB_PROJECT_ID=$PROJECT_ID \
  -p $PORT:8085 \
  --entrypoint /bin/sh \
  google/cloud-sdk:latest \
  -c "
    echo 'Starting emulator...';
    gcloud beta emulators pubsub start --host-port=0.0.0.0:8085 &
    sleep 5;
    export CLOUDSDK_AUTH_DISABLE_CREDENTIALS=1;
    export PUBSUB_EMULATOR_HOST=localhost:8085;
    export PUBSUB_PROJECT_ID=$PROJECT_ID;
    gcloud config set project $PROJECT_ID;
    until curl -s http://localhost:8085/v1/projects/$PROJECT_ID/topics > /dev/null; do sleep 1; done;
    curl -X PUT http://localhost:8085/v1/projects/$PROJECT_ID/topics/$TOPIC_ID;
    tail -f /dev/null
  "
