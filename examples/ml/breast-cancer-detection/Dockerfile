FROM pytorch/pytorch:0.4.1-cuda9-cudnn7-devel

RUN apt-get update && apt-get install -y libgl1-mesa-glx libglib2.0-0

WORKDIR /workspace
RUN git clone https://github.com/jimmywhitaker/breast_cancer_classifier.git /workspace
RUN pip install --upgrade pip && pip install -r requirements.txt

COPY . /workspace