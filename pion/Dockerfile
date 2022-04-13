FROM golang:1.17 AS build
WORKDIR /src
COPY ["pion", ""]
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh \
    && ./bin/golangci-lint --version
RUN ./bin/golangci-lint run main.go
RUN go build -o pion-echo main.go

FROM golang:1.17 AS final
WORKDIR /app
COPY --from=build /src/pion-echo .
ENTRYPOINT /app/pion-echo
