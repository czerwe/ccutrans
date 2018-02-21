FROM golang:latest as build

ARG GOOS=linux
ARG GOARCH=amd64
ENV CGO_ENABLED=0

RUN mkdir /go/src/ccutrans
COPY ccutrans.go /go/src/ccutrans/ccutrans.go
COPY ccuprocessing /go/src/ccutrans/ccuprocessing

WORKDIR /go/src/ccutrans

RUN go get -d .

RUN go build ccutrans


FROM scratch

COPY --from=build /go/src/ccutrans/ccutrans /ccutrans

ENV LISTENPORT 4040

EXPOSE 4040

ENTRYPOINT ["/ccutrans"]