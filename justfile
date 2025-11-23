default:
  just --list

rabbit:
  #!/usr/bin/env sh
  CONTAINER_NAME="rabbitmq"

  if docker container inspect "$CONTAINER_NAME" > /dev/null 2>&1; then
    # in case it was stopped
    docker start rabbitmq
  else
    # latest RabbitMQ 4 with management UI
    docker run --detach --rm --name "$CONTAINER_NAME" -p 5672:5672 -p 15672:15672 rabbitmq:4-management
  fi

test: rabbit
  go test ./...

run: rabbit
  go run . start

stop:
  docker stop rabbitmq
