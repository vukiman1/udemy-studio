# Translation style guide — "Fundamentals of Backend Engineering" (Hussein Nasser) EN → VI subtitles

## Audience & voice
- Readers are Vietnamese developers who already understand every English technical term. The goal is subtitles that read the way Vietnamese devs actually talk at work ("Vietlish" is welcome), not a formal textbook translation.
- Instructor speaks casually in first person. Use "mình" for "I", "bạn" / "các bạn" for "you", "chúng ta" for inclusive "we". Keep his tone: energetic, conversational, a bit opinionated.
- Concise: subtitles are read in seconds. Drop or lighten filler ("you know", "right?", "so yeah", "uh", stutters like "to to") — keep "đúng không?" / "ok" sparingly where it carries the rhythm.

## Keep in English (do NOT translate)
- Any term a dev would naturally say in English: request, response, client, server, backend, frontend, connection, socket, thread, process, kernel, user space, queue, buffer, packet, segment, frame, stream, header, payload, body, port, latency, throughput, bandwidth, handshake, timeout, retry, cache, proxy, reverse proxy, load balancer, sidecar, pub/sub, push, polling, long polling, short polling, event, callback, promise, blocking/non-blocking, sync/async, multiplexing, head-of-line blocking, congestion control, flow control, window, ACK, retransmission, encrypt/encryption, certificate, public key/private key, symmetric/asymmetric, hash, token, session, cookie, stateful/stateless, deploy, scale, scale out, idempotent, workload, pattern, design pattern, trade-off, overhead, bottleneck, fire-and-forget, broker, consumer, producer, subscriber, topic, schema, serialization, latency-sensitive, real-time.
- All protocol/product/tech names and acronyms: TCP, UDP, IP, ICMP, HTTP/1.1, HTTP/2, HTTP/3, QUIC, TLS, SSL, DNS, gRPC, WebSockets, WebRTC, SDP, STUN, TURN, ICE, NAT, MTU, TTL, SNI, ALPN, JWT, JSON, protobuf (protocol buffers), REST, SOAP, RPC, API, SQL, Postgres, MySQL, Redis, RabbitMQ, Kafka, NGINX, HAProxy, Envoy, Node.js, JavaScript, Chrome, Linux, epoll, io_uring, OSI, Layer 4/Layer 7.
- Code, commands, function names, URLs, numbers with units.
- Translate normal words normally ("gửi", "nhận", "chậm", "nhanh", "vấn đề", "ví dụ"...). Mixing like "client gửi request lên server rồi chờ response" is exactly the target style.

## Fix speech-recognition errors (captions are auto-generated)
Use context to restore the intended term. Known errors in this course:
- "back end"/"front end" → backend/frontend
- "STB two"/"HTTP two" → HTTP/2; "STB three" → HTTP/3; "STB" in HTTP contexts → HTTP
- "SDB" in WebRTC contexts → SDP; "SDB" in UDP/TCP error-handling contexts may be "TCP" — decide by context
- "GOPAC", "GRP C", "G RPC", "our gRPC" → gRPC
- "Jason" → JSON; "web RTC" → WebRTC; "web sockets" → WebSockets
- "TV library" / "TV" in gRPC/RPC contexts → likely "RPC"/"client library" — decide by context
- "quick" meaning the protocol → QUIC (but "real quick" = nhanh thôi)
If unsure what was meant, translate the most plausible technical meaning; never invent content.

## Hard format rules (the merge script rejects violations)
- Input lines are `<index>\t<text>`. Output exactly one line per input line: `<index>\t<vietnamese text>`, same indices, same order, tab separator, no line breaks inside text, never empty.
- Sentences often span several cues. Translate so each cue carries roughly its own portion of the meaning; you may shift a word or two between adjacent cues to keep Vietnamese word order natural, but every cue must keep some text.
- Don't merge, split, add or drop cues.
