FROM golang:1.17
RUN apt-get update
RUN apt-get -y install curl gnupg
RUN curl -sL https://deb.nodesource.com/setup_14.x  | bash -
RUN apt-get -y install nodejs
COPY test.sh /test/test.sh
WORKDIR /test
CMD ./test.sh
