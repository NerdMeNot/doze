---
title: "Recipes — SQS (queues)"
---


A ground-up, pure-Go SQS built into doze. It speaks both wire protocols (modern
AWS JSON 1.0 used by current SDKs and the legacy Query/XML), persists to disk, and
supports visibility timeouts, long polling, message attributes, FIFO, and
dead-letter redrive. Your SDK code talks to it unchanged.

- [Standard queue](#standard-queue)
- [FIFO queue](#fifo-queue-ordering--dedup)
- [Dead-letter + redrive](#dead-letter-queue--redrive)
- [Message attributes](#message-attributes)
- [Wire it into an app](#wire-it-into-an-app)
- [Common operations](#common-operations)

## Standard queue

```hcl
sqs "jobs" {
  queue "emails" {
    visibility_timeout = "30s"
    delay              = "0s"
    retention          = "96h"     # 4 days (Go durations or bare seconds)
    wait_time          = "10s"     # default long-poll wait
    max_message_size   = 262144
  }
}
```

```sh
# SQS listens on the explicit port you declared; set the endpoint + dummy creds:
export AWS_ENDPOINT_URL_SQS=http://127.0.0.1:9200
export AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_REGION=us-east-1
url=$(aws sqs get-queue-url --queue-name emails --query QueueUrl --output text)
aws sqs send-message    --queue-url "$url" --message-body "hello"
aws sqs receive-message --queue-url "$url" --wait-time-seconds 5
```

Durations use Go syntax (`30s`, `5m`, `12h`) or bare seconds — days aren't a unit,
so use hours (`96h` = 4 days).

## FIFO queue (ordering + dedup)

FIFO names must end in `.fifo`.

```hcl
sqs "orders" {
  queue "orders.fifo" {
    fifo                = true
    content_based_dedup = true     # dedupe by body hash within a 5-min window
  }
}
```

```sh
url=$(aws sqs get-queue-url --queue-name orders.fifo --query QueueUrl --output text)
aws sqs send-message --queue-url "$url" --message-body "o1" --message-group-id g
aws sqs send-message --queue-url "$url" --message-body "o1" --message-group-id g  # deduped away
```

Per-group ordering is preserved; while one message in a group is in flight, the
next in that group waits (until you delete it or its visibility lapses).

## Dead-letter queue + redrive

After `max_receive_count` receives without a delete, a message moves to the
dead-letter queue — exactly how you'd debug a "poison" message.

```hcl
sqs "work" {
  queue "tasks"     { visibility_timeout = "5s" }
  queue "tasks-dlq" {}
  redrive "tasks" {
    dead_letter       = "tasks-dlq"
    max_receive_count = 5
  }
}
```

```sh
# After 5 failed receives, inspect what got parked:
dlq=$(aws sqs get-queue-url --queue-name tasks-dlq --query QueueUrl --output text)
aws sqs receive-message --queue-url "$dlq"
```

## Message attributes

```sh
aws sqs send-message --queue-url "$url" --message-body "hi" \
  --message-attributes '{"kind":{"DataType":"String","StringValue":"signup"}}'
aws sqs receive-message --queue-url "$url" --message-attribute-names All
```

doze computes the AWS attribute MD5, so SDK checksum validation passes.

## Wire it into an app

Set the endpoint; credentials/region come from your env (exported as above, or
injected via a `process` block).

**Go:** `o.BaseEndpoint = aws.String(os.Getenv("AWS_ENDPOINT_URL_SQS"))`
**Node v3:** `new SQSClient({ endpoint: process.env.AWS_ENDPOINT_URL_SQS })`
**boto3:** `boto3.client("sqs", endpoint_url=os.environ["AWS_ENDPOINT_URL_SQS"])`

```sh
doze run -- ./worker        # processes the queue, backends guaranteed up
```

## Common operations

```sh
aws sqs get-queue-attributes --queue-url "$url" --attribute-names All
aws sqs purge-queue          --queue-url "$url"                    # empty it
aws sqs change-message-visibility --queue-url "$url" \
  --receipt-handle "$h" --visibility-timeout 0                     # release back immediately
```

Messages persist across reaps and restarts, and long polling is event-driven — a
send wakes a waiting receiver right away.
