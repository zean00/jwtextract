FROM golang:1.14.0 as builder

RUN mkdir -p /go/src/github.com/zean00/jwtextract
COPY ./main.go /go/src/github.com/zean00/jwtextract/main.go
RUN	cd /go/src/github.com/zean00/jwtextract && \
	env GOOS=linux GOARCH=amd64 go build -a -buildmode=plugin -trimpath -o jwextract.so
	
FROM genesix/krakend:latest
COPY --from=builder /go/src/github.com/zean00/jwtextract/jwextract.so /
