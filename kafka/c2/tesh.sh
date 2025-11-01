echo -n "Placeholder request" | nc -v localhost 9092 | hexdump -C

# 00 00 00 00  // message_size:   0 (any value works)
# 00 00 00 07  // correlation_id: 7