# Proxy NATS

This implementation helps us to take responsibility of balancing.

You should make a service and use server implementation together with Ping, as it exists in test.
You should take client, and embed it into your client application. We should subscribe on received subject.
Server will generate and send subject to subscribe.
Server will ping client app, and remove it from queue, if ping failed.