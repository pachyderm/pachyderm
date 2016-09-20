# Our image must derive from pachyderm/job-shim, otherwise Pachyderm won't be
# able to expose data to it.
FROM pachyderm/job-shim

# Install opencv and matplotlib.
RUN apt-get update \
	&& apt-get upgrade -y \
	&& apt-get install -y unzip wget build-essential \
		cmake git pkg-config libswscale-dev \
		python3-dev python3-numpy python3-tk \
		libtbb2 libtbb-dev libjpeg-dev \
		libpng-dev libtiff-dev libjasper-dev \
		bpython python3-pip libfreetype6-dev
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
	&& rm 3.1.0.zip
RUN sudo pip3 install matplotlib

# Add our own code.
ADD edges.py /edges.py
