# use appropriate image to start with
FROM python:slim

# Build time variables set from makefile
ARG PIPELINE_HOME
ARG PIPELINE_INPUT
ARG PIPELINE_OUTPUT
ARG SRC

# Environmental variables for input and output folders
ENV PIPELINE_HOME $PIPELINE_HOME
ENV PIPELINE_INPUT=$PIPELINE_INPUT PIPELINE_OUTPUT=$PIPELINE_OUTPUT

# we make some extra folders for testing purposes
RUN mkdir -p $PIPELINE_HOME

# copy over any relevant worker files and make appropriate changess
COPY $SRC/* $PIPELINE_HOME/
