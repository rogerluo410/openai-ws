FROM alpine:3.15

# No make in alpine  
# RUN make build-linux64

WORKDIR /var/app
COPY ./bin/openai-ws /var/app
COPY ./bin/config.yml /var/app

EXPOSE 8080 8090

ENTRYPOINT [ "./openai-ws" ]
