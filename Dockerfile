FROM quay.io/opsee/vinz:latest

ENV POSTGRES_CONN ""
ENV SPANX_ADDRESS ""
ENV SPANX_CERT="cert.pem"
ENV SPANX_CERT_KEY="key.pem"
ENV APPENV ""

RUN apk add --update bash ca-certificates curl
RUN curl -Lo /opt/bin/migrate https://s3-us-west-2.amazonaws.com/opsee-releases/go/migrate/migrate-linux-amd64 && \
    chmod 755 /opt/bin/migrate
RUN curl -Lo /opt/bin/ec2-env https://s3-us-west-2.amazonaws.com/opsee-releases/go/ec2-env/ec2-env && \
    chmod 755 /opt/bin/ec2-env

COPY run.sh /
COPY target/linux/amd64/bin/* /
COPY migrations /migrations
COPY cert.pem /
COPY key.pem /

EXPOSE 9095
CMD ["/spanx"]
