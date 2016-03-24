FROM alpine:3.3

RUN apk add --update bash ca-certificates curl
RUN mkdir -p /opt/bin && \
		curl -Lo /opt/bin/s3kms https://s3-us-west-2.amazonaws.com/opsee-releases/go/vinz-clortho/s3kms-linux-amd64 && \
    chmod 755 /opt/bin/s3kms && \
    curl -Lo /opt/bin/migrate https://s3-us-west-2.amazonaws.com/opsee-releases/go/migrate/migrate-linux-amd64 && \
    chmod 755 /opt/bin/migrate

ENV POSTGRES_CONN ""
ENV SPANX_ADDRESS ""
ENV SPANX_CERT="cert.pem"
ENV SPANX_CERT_KEY="key.pem"
ENV APPENV ""

COPY run.sh /
COPY target/linux/amd64/bin/* /
COPY migrations /migrations
COPY cert.pem /
COPY key.pem /

EXPOSE 9095
CMD ["/spanx"]
