**A. Lexical / Natural Sorting for Keys or Sorted Sets**

Redis sorts keys and sorted sets byte-wise, not naturally.
You can implement a version of ZRANGEBYLEX that supports:
Case-insensitive sorting
Unicode-aware / accent-insensitive sorting

**B. Simple Expiry / TTL System**

Redis supports key expiration with automatic deletion.
You could implement:
TTL per key
Background cleanup goroutine
Why it’s good:
Teaches about time-based operations, goroutines, and concurrency safety.
