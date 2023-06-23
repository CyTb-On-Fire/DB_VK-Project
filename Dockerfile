FROM golang:1.20-alpine3.18 as build
COPY . /project
WORKDIR /project
RUN mkdir -p build && go build -o build forum/cmd/main

FROM postgres:15.3-alpine3.18 as main
COPY --from=build /project/build/main main
COPY --from=build /project/scripts/run.sh /docker-entrypoint-initdb.d/
COPY scripts/db/db.sql /docker-entrypoint-initdb.d/

ENV POSTGRES_PASSWORD=12345
ENV POSTGRES_DB=forumservice

EXPOSE 5000
