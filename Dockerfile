FROM golang
LABEL maintainer="Henrique Vicente <henrique.vicente@liferay.cloud>"

ENV KUBEAPPLY_LATEST_VERSION="v1.13.3"

ADD https://storage.googleapis.com/kubernetes-release/release/${KUBEAPPLY_LATEST_VERSION}/bin/linux/amd64/kubectl /bin/kubectl
RUN [ "chmod", "+x", "/bin/kubectl" ]
COPY . /go/src/github.com/henvic/kubeapply

RUN [ "go", "build", "-o", "/bin/kubeapply", "/go/src/github.com/henvic/kubeapply/cmd/server" ]

EXPOSE 8080
ENTRYPOINT /bin/kubeapply
