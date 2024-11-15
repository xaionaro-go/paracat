# ParaCat: A SIMO/MISO forwarder for high reliability/throughput

In crowded Internet, all connections are not reliable. To minimize jitter and packet loss, we can send it through different routes simultaneously then get redundancy.

## Structure

```mermaid
flowchart LR
    C[UDP Client]
    IN[Inbound]
    OUT[Outbound]
    R0(Relay Server #0)
    R1(Relay Server #1)
    Rn(Relay Server #n)
    S(UDP Server)

    C --> IN

    IN -->|Raw TCP| R0
    IN -->|Raw UDP| R0
    IN -->|SOCKS5| R1
    IN -->|Others| Rn
    IN -->|Raw UDP| OUT

    R0 --> OUT
    R1 --> OUT
    Rn --> OUT

    OUT --> S
```

## TODO

- [ ] Remove unused UDP connections
- [ ] Optimize delay
- [ ] Re-connect after EOF
- [ ] Round-robin mode
