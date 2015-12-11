FROM golang
ADD GeoLite2-City.mmdb GeoLite2-City.mmdb
ADD ./vendor /go/src/github.com/klauspost/InterviewAssignment/vendor
ADD ./cmd /go/src/github.com/klauspost/InterviewAssignment/cmd
ADD ./traffic /go/src/github.com/klauspost/InterviewAssignment/traffic
ENV GO15VENDOREXPERIMENT 1

RUN go install github.com/klauspost/InterviewAssignment/cmd/importlogs

ENTRYPOINT ["importlogs", "-geodb=GeoLite2-City.mmdb", "/import/NASA_access_log_Jul95.gz", "/import/NASA_access_log_Aug95.gz"]
VOLUME /import
