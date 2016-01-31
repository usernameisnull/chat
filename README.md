# Tinode Instant Messaging Server

Instant messaging server. Backend in pure [Go](http://golang.org) ([Affero GPL 3.0](http://www.gnu.org/licenses/agpl-3.0.en.html)), client-side binding in Java for Android and Javascript ([Apache 2.0](http://www.apache.org/licenses/LICENSE-2.0)), persistent storage [RethinkDB](http://rethinkdb.com/), JSON over websocket (long polling is also available). No UI components other than demo apps. Tinode is meant as a replacement for XMPP.

Version 0.5. This is alpha-quality software. Bugs should be expected. Follow [instructions](INSTALL.md) to install and run. Read [API documentation](API.md).

A demo is (usually) available at http://api.tinode.co/x/example-react-js/ ([source](https://github.com/tinode/example-react-js/)). Login as one of `alice`, `bob`, `carol`, `dave`, `frank`. Password is `<login>123`, e.g. login for `alice` is `alice123`.


## Why?

[XMPP](http://xmpp.org/) is a mature specification with support for a very broad spectrum of use cases developed long before mobile became important. As a result most (all?) known XMPP servers are difficult to adapt for the most common use case of a few people messaging each other from mobile devices. Tinode is an attempt to build a modern replacement for XMPP/Jabber focused on a narrow use case of instant messaging between humans with emphasis on mobile communication.

## Features

### Supported

* One-on-one messaging.
* Groups (topics) with up to 32 members where every member's access permissions are managed individually.
* Topic access control with separate permissions for various actions.
* Server-generated presence notifications for people, topics.
* Persistent message store.
* Javascript bindings with no dependencies.
* Websocket & long polling transport.
* JSON wire protocol.
* Message delivery status: server-generated delivery to server, client-generated received and read notifications.
* Basic support for client-side message caching.
* Ability to block unwanted communication on the server.

### Planned

* iOS client bindings.
* Android Java bindings (Current implementation is incomplete, dependencies: [jackson](https://github.com/FasterXML/jackson), [android-websockets](https://github.com/codebutler/android-websockets))
* Mobile push notification hooks.
* Groups (topics) with unlimited number of members with bearer token access control.
* Clustering.
* Federation.
* Different levels of message persistence (from strict persistence to store until delivered to purely ephemeral messaging).
* Support for binary wire protocol.
* User search/discovery.
* Anonymous users.
* Support for other SQL and NoSQL backends.
* Pluggable authentication.
* Plugins.
