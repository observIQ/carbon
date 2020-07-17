FROM ubuntu:bionic

RUN mkdir -p /carbon_home/plugins
ENV CARBON_HOME=/carbon_home
RUN echo "pipeline:\n" >> /carbon_home/config.yaml
RUN apt-get update && apt-get install -y systemd ca-certificates

COPY ./artifacts/carbon_linux_amd64 /carbon_home/carbon
ENTRYPOINT /carbon_home/carbon \
  --config /carbon_home/config.yaml \
  --database /carbon_home/carbon.db \
  --plugin_dir /carbon_home/plugins
