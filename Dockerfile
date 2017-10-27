FROM golang:1.9
WORKDIR /go/src/github.com/manifoldco/kubernetes-credentials
COPY Makefile .
COPY Gopkg.* ./
RUN make bootstrap
RUN make vendor
COPY . ./
RUN make bin/controller

FROM centurylink/ca-certs
COPY --from=0 /go/src/github.com/manifoldco/kubernetes-credentials/bin/controller .
CMD ["./controller"]
