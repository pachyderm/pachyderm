FROM alpine:3.3
COPY sudoers /etc/
RUN apk update && \
    apk add sudo && \
    adduser -u 1000 -G wheel -D alpine && \
    rm -rf /var/cache/apk/*
USER alpine

ENV NVIDIA_RUNNER NVIDIA-Linux-x86_64-375.51.run
RUN wget http://us.download.nvidia.com/XFree86/Linux-x86_64/375.51/${NVIDIA_RUNNER} && \
	chmod +x ${NVIDIA_RUNNER}
ADD install-nvidia-driver.sh .
CMD ./install-nvidia-driver.sh ${NVIDIA_RUNNER}
