
FROM nvidia/cuda:11.7.1-devel-ubuntu20.04

ENV TZ=Asia/Tokyo
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN apt-get update
RUN apt-get install -y python3 python3-pip
RUN pip3 install torch torchvision
RUN apt-get update && apt-get install -y libsndfile1 ffmpeg
RUN pip3 install Cython
RUN pip3 install nemo_toolkit['nlp']

WORKDIR /work

COPY nemo.py /work/

ENV LIBRARY_PATH /usr/local/cuda/lib64/stubs