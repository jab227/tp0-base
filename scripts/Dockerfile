FROM debian:latest
WORKDIR scripts

COPY ./scripts/nc_test.sh ./nc_test.sh
RUN apt-get update -y &&\
    apt-get install netcat-traditional -y &&\
    chmod +x ./nc_test.sh
ENTRYPOINT ["/bin/bash"]
