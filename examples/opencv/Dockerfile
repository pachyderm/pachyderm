FROM ubuntu:20.04
# Fix timezone issue
ENV TZ=UTC
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Install opencv and matplotlib.
RUN export DEBIAN_FRONTEND=noninteractive; \
    export DEBCONF_NONINTERACTIVE_SEEN=true; \
    echo 'tzdata tzdata/Areas select Etc' | debconf-set-selections; \
    echo 'tzdata tzdata/Zones/Etc select UTC' | debconf-set-selections; \
    apt-get update -qqy \
    && apt-get install -qqy make git pkg-config libswscale-dev python3-dev \
    	python3-numpy python3-tk libtbb2 libtbb-dev libjpeg-dev libpng-dev \
    	libtiff-dev bpython python3-pip libfreetype6-dev wget unzip cmake \
    	sudo \
    && apt-get clean \
    && rm -rf /var/lib/apt

RUN cd \
	&& wget https://github.com/Itseez/opencv/archive/3.4.5.zip \
	&& unzip 3.4.5.zip \
	&& cd opencv-3.4.5 \
	&& mkdir build \
	&& cd build \
	&& cmake .. \
	&& make -j \
	&& make install \
	&& cd \
	&& rm 3.4.5.zip \
    && rm -rf opencv-3.4.5
RUN python3 --version && pip3 --version && sudo pip3 install matplotlib

# Add our own code.
ADD edges.py /edges.py
