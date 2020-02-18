
FROM amazon/aws-for-fluent-bit:latest

COPY bin/s3.so /fluent-bit/s3.so

COPY fluent-bit.conf /fluent-bit/etc/

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Optional Metrics endpoint
EXPOSE 2020

# Entry point
CMD /entrypoint.sh