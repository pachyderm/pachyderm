FROM alpine

ENV NVIDIA_RUNNER NVIDIA-Linux-x86_64-375.51.run
RUN wget http://us.download.nvidia.com/XFree86/Linux-x86_64/375.51/${NVIDIA_RUNNER} && \
	chmod +x ${NVIDIA_RUNNER}
ADD install-nvidia-driver.sh .
CMD ./install-nvidia-driver.sh
