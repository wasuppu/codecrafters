echo -n "00000023001200046f7fc66100096b61666b612d636c69000a6b61666b612d636c6904302e3100" | xxd -r -p | nc localhost 9092 | hexdump -C

# 00 00 00 00  // message_size:   0 (any value works)
# 6f 7f c6 61  // correlation_id: 1870644833