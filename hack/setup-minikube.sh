# YOU NEED TO CREATE A GITHUB CLASSIC TOKEN TO BE ABLE TO PULL IMAGES FROM Open-Digital-Twin GROUP PACKAGES, THEN RUN:

# echo YOUR_GITHUB_PAT | docker login ghcr.io -u YOUR_GITHUB_USERNAME --password-stdin

# Load KTWIN Containers into Minikube
minikube image load ghcr.io/open-digital-twin/ktwin-event-store:0.1
minikube image load ghcr.io/open-digital-twin/ktwin-mqtt-dispatcher:0.1
minikube image load ghcr.io/open-digital-twin/ktwin-cloud-event-dispatcher:0.1
minikube image load ghcr.io/open-digital-twin/ktwin-pole-service:0.1

# Development utilities
docker pull curlimages/curl:latest
minikube image load curlimages/curl:latest

