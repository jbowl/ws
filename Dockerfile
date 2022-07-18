FROM golang:1.18.3-alpine3.16 as build

# share pkg from private repo
#ARG GITLAB_USER
#ARG GITLAB_TOKEN

RUN apk add --no-cache git # Add git to alpine to get the commit and branch
#RUN apk add --update coreutils # Add GNU date to alpine to get the build date

WORKDIR /app
COPY . src/

RUN cd /app/src && go mod tidy

# embed git info via linker 
RUN cd /app/src && CGO_ENABLED=0 \
    go build -ldflags="-s -w"\
    #-ldflags="-s -w -X main.Commit=`git rev-parse --short HEAD` \
    #-X main.Date=`date -u --rfc-3339=seconds | sed -e 's/ /T/'` \
    #-X  main.Branch=`git symbolic-ref -q --short HEAD` " \
    -o /app/ws

# final stage
FROM alpine:3.16
RUN apk add --no-cache aws-cli

WORKDIR /app
COPY --from=build /app/src/templates /app/templates 
# local dev COPY --from=build /app/src/cert.pem /app/
COPY --from=build /app/ws /app/ 
COPY --from=build /app/src/start.sh /app/
RUN chmod +x /app/start.sh

ENV GO_ENV=production \
    PORT=50002 \ 
    BREWERY_ADDR=breweries.jbowl.dev:50051

CMD ["/app/start.sh"]
