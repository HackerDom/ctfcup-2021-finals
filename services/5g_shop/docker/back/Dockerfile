FROM ubuntu:20.04

RUN apt update \
    && apt install -y \
        libboost-dev \
        libboost-log-dev \
        libboost-program-options-dev \
        libpq-dev

RUN useradd --no-create-home --shell /bin/false --uid 1000 --user-group 5g_shop

COPY 5g_shop /

RUN chmod +x /5g_shop

USER 5g_shop

CMD ["/5g_shop", "--address", "0.0.0.0", "--port", "31337"]

