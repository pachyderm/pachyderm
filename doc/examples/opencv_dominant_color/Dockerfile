FROM ubuntu:14.04
WORKDIR /home

# OpenCV dynamic link dependancies
RUN sudo apt-get update && sudo apt-get install -y libjpeg-dev libpng12-dev
# OpenCV C++ Code
ADD dominantColor /home
# Golang wrapper
ADD dominant_color /home
