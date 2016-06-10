FROM golang:1.6
WORKDIR /go/src/github.com/warmans/stressfaktor-api
ADD . .
ENV GLIDE_VERSION 0.8.3
RUN apt-get update && apt-get install -y unzip --no-install-recommends && rm -rf /var/lib/apt/lists/*
RUN curl \
	-fsSL "https://github.com/Masterminds/glide/releases/download/${GLIDE_VERSION}/glide-${GLIDE_VERSION}-linux-amd64.zip" -o glide.zip \
	&& unzip glide.zip  linux-amd64/glide \
	&& mv linux-amd64/glide /usr/local/bin \
	&& rm -rf linux-amd64 \
	&& rm glide.zip
RUN make build
ENTRYPOINT ["stressfaktor-api"]
EXPOSE 8080
