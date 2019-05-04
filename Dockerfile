FROM golang:1.12.4 as builder
LABEL maintainer="Henrique Vicente <henrique.vicente@liferay.cloud>"

# check newest version at https://storage.googleapis.com/kubernetes-release/release/stable.txt
ENV KUBECTL_LATEST_VERSION="v1.14.1"

ADD https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_LATEST_VERSION}/bin/linux/amd64/kubectl /bin/kubectl
COPY . /go/src/github.com/henvic/kubeapply

# disable CGO so we can use multi-stage with alpine. Otherwise, this error happens:
# standard_init_linux.go:207: exec user process caused "no such file or directory"
ENV CGO_ENABLED="0"

RUN [ "go", "build", "-o", "/bin/kubeapply", "/go/src/github.com/henvic/kubeapply/cmd/server" ]

FROM alpine
RUN apk add curl
RUN apk --no-cache add ca-certificates

COPY --from=builder /bin/kubeapply /bin
COPY --from=builder /bin/kubectl /bin
RUN [ "chmod", "+x", "/bin/kubectl" ]
RUN [ "chmod", "+x", "/bin/kubeapply" ]

EXPOSE 9000
ENTRYPOINT [ "/bin/kubeapply", "-addr=0.0.0.0:9000" ]
