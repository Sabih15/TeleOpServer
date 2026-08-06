# TeleOpServer — Issues & Resolutions

> A log of non-obvious problems encountered during development, their root causes, and how they were resolved.

---

## Issue 1 — TimescaleDB extension not available on startup

**Error**
```
ERROR: extension "timescaledb" is not available (SQLSTATE 0A000)
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE
failed to initialize application
```

**Symptom**  
App failed to start immediately after `docker-compose up`. The error pointed to `postgres.go:18` where the extension is created.

**Root cause**  
Two PostgreSQL instances were listening on port 5432 simultaneously:
- A **native Windows PostgreSQL installation** (PID 7064)
- The **Docker container** (via Docker backend, PID 19660)

When the app connected to `localhost:5432`, it hit the native installation which had no TimescaleDB extension installed. The Docker container with the correct `timescale/timescaledb:latest-pg16` image was never reached.

**Diagnosis**
```powershell
netstat -ano | findstr ":5432"
# Two processes on port 5432 confirmed the conflict
```

**Resolution**  
Changed the Docker container's port mapping from `5432:5432` to `5433:5432` in `docker-compose.yml`, and updated `DATABASE_URL` in `.env` to use port `5433`. The native Postgres installation on 5432 is left untouched.

```yaml
# docker-compose.yml
ports:
  - "5433:5432"   # was "5432:5432"
```

```
# .env
DATABASE_URL=host=localhost port=5433 dbname=teleopserver user=postgres password=postgres1 sslmode=disable
```

---

## Issue 2 — MQTT queued messages dropped on startup

**Symptom**  
With `CleanSession: false` and QoS 1 configured, messages published while the server was offline were visible in the RabbitMQ management UI queue. However, on server restart those messages disappeared from the queue without being processed — no log output, no DB records.

**Root cause**  
A race condition in the startup order:

1. `mqttclient.New()` called `Connect()` immediately on construction
2. The broker detected the client was back online and began delivering queued messages right away
3. `OnConnectHandler` fired — but `c.subs` was **empty** because `consumer.Start()` had not yet called `Subscribe()`
4. Paho received the messages with no registered handler → auto-acknowledged and discarded them
5. `consumer.Start()` registered the handler moments later — too late

**Resolution**  
Removed `Connect()` from `mqttclient.New()`. The client is now constructed but not connected. `consumer.Start()` calls `Subscribe()` first (populating `c.subs`), then calls `Connect()`. By the time the broker delivers queued messages, all handlers are guaranteed to be in place.

```go
// consumer.go — correct order
func (c *Consumer) Start(ctx context.Context) error {
    c.mqtt.Subscribe(commandTopic, handler)  // handler in c.subs FIRST
    if err := c.mqtt.Connect(); err != nil { // THEN connect
        return err
    }
    <-ctx.Done()
    c.mqtt.Disconnect()
    return nil
}
```

---

## Issue 3 — Potential deadlock in OnConnect handler

**Symptom**  
Not a runtime crash, but a latent deadlock identified during code review.

**Root cause**  
The original `OnConnectHandler` acquired `c.mu.Lock()` and then called `inner.Subscribe()` — a blocking network operation — while still holding the lock. Any concurrent call to `client.Subscribe()` would attempt to acquire the same mutex and deadlock.

```go
// original — deadlock risk
SetOnConnectHandler(func(inner mqtt.Client) {
    c.mu.Lock()
    defer c.mu.Unlock()           // held for entire duration
    for topic, sub := range c.subs {
        inner.Subscribe(topic, 1, ...) // blocking network call under lock
    }
})
```

**Resolution**  
Copy the subscriptions into a local snapshot under the lock, release the lock, then perform the network calls outside it.

```go
// fixed — snapshot then subscribe
SetOnConnectHandler(func(inner mqtt.Client) {
    c.mu.Lock()
    subs := make(map[string]subscription, len(c.subs))
    for k, v := range c.subs { subs[k] = v }
    c.mu.Unlock()                  // released before network calls

    for topic, sub := range subs {
        sub := sub
        inner.Subscribe(topic, 1, ...)
    }
})
```

---

## Issue 4 — App crashes if MQTT consumer fails to start

**Symptom**  
Not a runtime crash observed, but a code review finding. If `consumer.Start()` returned an error, `log.Fatal()` would call `os.Exit(1)` immediately — bypassing HTTP server shutdown, database connection cleanup, and MQTT disconnect.

**Root cause**  
```go
// original — kills the entire process on subscribe failure
go func() {
    if err := consumer.Start(ctx); err != nil {
        log.Fatal().Err(err).Msg("mqtt consumer stopped")
    }
}()
```

`log.Fatal()` is equivalent to `log.Error()` + `os.Exit(1)`. There is no deferred cleanup, no graceful HTTP drain, and no clean MQTT disconnect.

**Resolution**  
Changed to `log.Error()` so a subscribe failure is logged but the HTTP server keeps running. History query endpoints remain available even without a live MQTT connection.

```go
go func() {
    if err := consumer.Start(ctx); err != nil {
        log.Error().Err(err).Msg("mqtt consumer failed to start, commands will not be ingested")
    }
}()
```

---

## Issue 5 — No graceful shutdown

**Symptom**  
Pressing `Ctrl+C` killed the process immediately. In-flight HTTP requests were dropped, the MQTT client did not send a clean disconnect to the broker, and database connections were closed ungracefully.

**Root cause**  
`server.Run()` called `http.ListenAndServe()` directly with no signal handling. The process received `SIGINT` and terminated without cleanup.

**Resolution**  
Added `signal.NotifyContext` in `main.go` to capture `SIGINT` / `SIGTERM`. The context is threaded through to the HTTP server and MQTT consumer. On shutdown:

1. `server.Run(ctx)` calls `srv.Shutdown()` with a 10-second timeout — drains in-flight requests
2. `consumer.Start(ctx)` unblocks from `<-ctx.Done()` and calls `mqtt.Disconnect()`

```go
// main.go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

app, err := wire.InitializeApp(ctx)
// ...
app.Run(ctx)  // blocks until Ctrl+C, then shuts down cleanly
```
