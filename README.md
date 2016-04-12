# go-slackgw

Yet Another Slack Gateway

# Features

* HTTP interface to allow easy integration within trusted environments
* RTM interface that allows you to queue/process incoming messages

# HTTP interface

## Send a message

First, start the server (127.0.0.1:4979):

```
slackgw \
    -token=/path/to/tokenfile
```

Then send a POST message:

```
curl -XPOST http://slackgw:4979/post -d "channel=#updates&message=test" 
```

This is great for your organization/company wide monitoring tools and such, especially when you have lots of tools that may want to post messages to slack, and you are too lazy creating tokens for each bot you have.

Do NOT open this up for the wider internet.

# RTM interface

## Queue incoming message events to Google PubSub

```
slackgw \
    -rtm=gpubsub-forward \
    -event=MessageEvent \
    -topic=projects/:project_id:/topics/:topic: \
    -token=/path/to/tokenfile
```

Or, embed in some other slack integration of yours:

```go
  hctx := context.Background()
  httpcl, err := google.DefaultClient(hctx, pubsub.PubsubScope)
  if err != nil {
    return fmt.Errorf("Failed to create default oauth client: %s", err)
  }
  pubsubsvc, err := pubsub.New(httpcl)
  if err != nil {
    return fmt.Errorf("Failed to create pubsub client: %s", err)
  }
  s := slackgw.New()
  s.StartSlack(token)
  s.StartRTM(gcp.NewPubsubForwarder(pubsubsvc, topic, slackgw.MessageEvent))

  // other initializations follow...
```
