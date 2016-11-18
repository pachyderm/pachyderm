FROM ubuntu:14.04

# Install opencv and matplotlib.
RUN apt-get update \
	&& apt-get upgrade -y \
	&& apt-get install -y unzip wget build-essential \
		cmake git pkg-config libswscale-dev \
		python3-dev python3-numpy python3-tk \
		libtbb2 libtbb-dev libjpeg-dev \
		libpng-dev libtiff-dev libjasper-dev \
		bpython python3-pip libfreetype6-dev \
    && apt-get clean \
    && rm -rf /var/lib/apt
RUN cd \
	&& wget https://github.com/Itseez/opencv/archive/3.1.0.zip \
	&& unzip 3.1.0.zip \
	&& cd opencv-3.1.0 \
	&& mkdir build \
	&& cd build \
	&& cmake .. \
	&& make -j \
	&& make install \
	&& cd \
	&& rm 3.1.0.zip \
    && rm -rf opencv-3.1.0
RUN sudo pip3 install matplotlib

# Add our own code.
ADD edges.py /edges.py
