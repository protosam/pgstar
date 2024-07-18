FROM golang:1.22 AS gobuilder
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY . /app
RUN CGO_ENABLED=0 go build -o /pgstar github.com/protosam/pgstar/cli

FROM scratch
COPY --from=gobuilder /pgstar /pgstar
CMD [ "/pgstar" ]
ENTRYPOINT [ "/pgstar" ]
