FROM alpine:latest

RUN mkdir /app

COPY urlShotenerApp /app

CMD [ "/app/urlShotenerApp" ]