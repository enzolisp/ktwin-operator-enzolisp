# Pull KTWIN Containers
docker pull dev.local/open-digital-twin/ktwin-event-store:0.1
docker pull dev.local/open-digital-twin/ktwin-mqtt-dispatcher:0.1
docker pull dev.local/open-digital-twin/ktwin-cloud-event-dispatcher:0.1
docker pull dev.local/open-digital-twin/ktwin-pole-service:0.1

# Load KTWIN Containers into Kind
kind load docker-image dev.local/open-digital-twin/ktwin-event-store:0.1
kind load docker-image dev.local/open-digital-twin/ktwin-mqtt-dispatcher:0.1
kind load docker-image dev.local/open-digital-twin/ktwin-cloud-event-dispatcher:0.1
kind load docker-image dev.local/open-digital-twin/ktwin-pole-service:0.1

# Development utilities
docker pull curlimages/curl:latest
kind load docker-image curlimages/curl:latest