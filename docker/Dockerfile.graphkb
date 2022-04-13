FROM golang:1.18

ARG USER_ID
ARG GROUP_ID

RUN groupadd --gid ${GROUP_ID} dev && \
    useradd --uid ${USER_ID} -s /bin/bash -m -g dev dev

USER dev