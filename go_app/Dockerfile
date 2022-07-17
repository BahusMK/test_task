FROM golang:alpine
COPY app  /opt/app
WORKDIR /opt/app
RUN cd /opt/app && \
    go env -w GO111MODULE=on && \
    go mod init app && \
    go mod tidy && \
    go build -o ./bin_app
EXPOSE 8080
CMD [ "/opt/app/bin_app"]