FROM golang:1.20-alpine3.18 as build
COPY . /project
WORKDIR /project
RUN mkdir -p build && go build -o build DBProject/cmd/app

FROM postgres:15.3-alpine3.18 as main
COPY --from=build /project/build/app main
COPY --from=build /project/scripts/run.sh /docker-entrypoint-initdb.d/
COPY db/db.sql /docker-entrypoint-initdb.d/
COPY scripts/ /docker-entrypoint-initdb.d/

ENV POSTGRES_PASSWORD=12345
ENV POSTGRES_DB=forumservice

RUN chmod 777 /docker-entrypoint-initdb.d/run.sh

EXPOSE 5000

