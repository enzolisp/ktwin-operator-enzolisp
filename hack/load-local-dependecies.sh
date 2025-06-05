# YOU NEED TO CREATE A GITHUB CLASSIC TOKEN TO BE ABLE TO PULL IMAGES FROM Open-Digital-Twin GROUP PACKAGES, THEN RUN:

# echo YOUR_GITHUB_PAT | docker login ghcr.io -u YOUR_GITHUB_USERNAME --password-stdin

# Pull KTWIN Containers
# docker pull dev.local/open-digital-twin/ktwin-event-store:0.1
# docker pull dev.local/open-digital-twin/ktwin-mqtt-dispatcher:0.1
# docker pull dev.local/open-digital-twin/ktwin-cloud-event-dispatcher:0.1
# docker pull dev.local/open-digital-twin/ktwin-pole-service:0.1

# Pull KTWIN Containers
docker pull ghcr.io/open-digital-twin/ktwin-event-store:0.1
docker pull ghcr.io/open-digital-twin/ktwin-mqtt-dispatcher:0.1
docker pull ghcr.io/open-digital-twin/ktwin-cloud-event-dispatcher:0.1
docker pull ghcr.io/open-digital-twin/ktwin-pole-service:0.1

# Load KTWIN Containers into Kind
kind load docker-image ghcr.io/open-digital-twin/ktwin-event-store:0.1
kind load docker-image ghcr.io/open-digital-twin/ktwin-mqtt-dispatcher:0.1
kind load docker-image ghcr.io/open-digital-twin/ktwin-cloud-event-dispatcher:0.1
kind load docker-image ghcr.io/open-digital-twin/ktwin-pole-service:0.1

# Development utilities
docker pull curlimages/curl:latest
kind load docker-image curlimages/curl:latest
