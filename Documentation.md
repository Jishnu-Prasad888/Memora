# Memora Redis-like Database Commands Documentation

## Table of Contents
- [String Commands](#string-commands)
- [List Commands](#list-commands)
- [Set Commands](#set-commands)
- [Hash Commands](#hash-commands)
- [Key Commands](#key-commands)
- [Server Commands](#server-commands)

## String Commands

### SET
Sets the string value of a key.

**Syntax:**
```
SET key value [EX seconds] [PX milliseconds]
```

**Arguments:**
- `key` - The key to set
- `value` - The string value
- `EX seconds` - Set expiry time in seconds (optional)
- `PX milliseconds` - Set expiry time in milliseconds (optional)

**Examples:**
```
> SET username "john_doe"
"OK"

> SET session_token "abc123" EX 3600
"OK"

> SET temp_data "value" PX 5000
"OK"
```

**Return:**
- `"OK"` on success

---

### GET
Gets the value of a key.

**Syntax:**
```
GET key
```

**Arguments:**
- `key` - The key to get

**Examples:**
```
> SET message "Hello World"
"OK"

> GET message
"Hello World"

> GET nonexistent
(nil)
```

**Return:**
- String value if key exists
- `(nil)` if key doesn't exist

---

### INCR
Increments the integer value of a key by 1.

**Syntax:**
```
INCR key
```

**Arguments:**
- `key` - The key to increment

**Examples:**
```
> SET counter 10
"OK"

> INCR counter
(integer) 11

> INCR counter
(integer) 12
```

**Return:**
- Integer value after increment
- Error if value is not an integer

---

### DECR
Decrements the integer value of a key by 1.

**Syntax:**
```
DECR key
```

**Arguments:**
- `key` - The key to decrement

**Examples:**
```
> SET counter 10
"OK"

> DECR counter
(integer) 9

> DECR counter
(integer) 8
```

**Return:**
- Integer value after decrement
- Error if value is not an integer

---

## List Commands

### LPUSH
Inserts values at the head (left) of a list.

**Syntax:**
```
LPUSH key value [value ...]
```

**Arguments:**
- `key` - The list key
- `value` - One or more values to push

**Examples:**
```
> LPUSH mylist "world"
(integer) 1

> LPUSH mylist "hello"
(integer) 2

> LRANGE mylist 0 -1
1) "hello"
2) "world"
```

**Return:**
- Integer length of list after push

---

### RPUSH
Inserts values at the tail (right) of a list.

**Syntax:**
```
RPUSH key value [value ...]
```

**Arguments:**
- `key` - The list key
- `value` - One or more values to push

**Examples:**
```
> RPUSH mylist "hello"
(integer) 1

> RPUSH mylist "world"
(integer) 2

> LRANGE mylist 0 -1
1) "hello"
2) "world"
```

**Return:**
- Integer length of list after push

---

### LPOP
Removes and returns the first element of a list.

**Syntax:**
```
LPOP key
```

**Arguments:**
- `key` - The list key

**Examples:**
```
> RPUSH mylist "one" "two" "three"
(integer) 3

> LPOP mylist
"one"

> LPOP mylist
"two"
```

**Return:**
- The popped value, or `(nil)` if list is empty

---

### RPOP
Removes and returns the last element of a list.

**Syntax:**
```
RPOP key
```

**Arguments:**
- `key` - The list key

**Examples:**
```
> RPUSH mylist "one" "two" "three"
(integer) 3

> RPOP mylist
"three"

> RPOP mylist
"two"
```

**Return:**
- The popped value, or `(nil)` if list is empty

---

### LLEN
Returns the length of a list.

**Syntax:**
```
LLEN key
```

**Arguments:**
- `key` - The list key

**Examples:**
```
> RPUSH mylist "a" "b" "c"
(integer) 3

> LLEN mylist
(integer) 3

> LLEN nonexistent
(integer) 0
```

**Return:**
- Integer length of the list

---

## Set Commands

### SADD
Adds members to a set.

**Syntax:**
```
SADD key member [member ...]
```

**Arguments:**
- `key` - The set key
- `member` - One or more members to add

**Examples:**
```
> SADD myset "hello"
(integer) 1

> SADD myset "world"
(integer) 1

> SADD myset "hello"
(integer) 0  # Already exists
```

**Return:**
- Integer number of members added (excluding duplicates)

---

### SREM
Removes members from a set.

**Syntax:**
```
SREM key member [member ...]
```

**Arguments:**
- `key` - The set key
- `member` - One or more members to remove

**Examples:**
```
> SADD myset "one" "two" "three"
(integer) 3

> SREM myset "one"
(integer) 1

> SREM myset "four"
(integer) 0  # Didn't exist
```

**Return:**
- Integer number of members removed

---

### SMEMBERS
Returns all members of a set.

**Syntax:**
```
SMEMBERS key
```

**Arguments:**
- `key` - The set key

**Examples:**
```
> SADD myset "apple" "banana" "cherry"
(integer) 3

> SMEMBERS myset
1) "apple"
2) "banana"
3) "cherry"

> SMEMBERS nonexistent
(empty array)
```

**Return:**
- Array of all set members

---

### SISMEMBER
Checks if a member exists in a set.

**Syntax:**
```
SISMEMBER key member
```

**Arguments:**
- `key` - The set key
- `member` - The member to check

**Examples:**
```
> SADD myset "hello"
(integer) 1

> SISMEMBER myset "hello"
(integer) 1

> SISMEMBER myset "world"
(integer) 0
```

**Return:**
- `1` if member exists
- `0` if member doesn't exist

---

## Hash Commands

### HSET
Sets field in a hash to a value.

**Syntax:**
```
HSET key field value [field value ...]
```

**Arguments:**
- `key` - The hash key
- `field` - The field name
- `value` - The field value

**Examples:**
```
> HSET user:1000 name "John" age 30
(integer) 2

> HSET user:1000 email "john@example.com"
(integer) 1
```

**Return:**
- Integer number of fields added

---

### HGET
Gets the value of a field in a hash.

**Syntax:**
```
HGET key field
```

**Arguments:**
- `key` - The hash key
- `field` - The field name

**Examples:**
```
> HSET user:1000 name "John"
(integer) 1

> HGET user:1000 name
"John"

> HGET user:1000 age
(nil)
```

**Return:**
- Field value, or `(nil)` if field doesn't exist

---

### HDEL
Deletes fields from a hash.

**Syntax:**
```
HDEL key field [field ...]
```

**Arguments:**
- `key` - The hash key
- `field` - One or more fields to delete

**Examples:**
```
> HSET user:1000 name "John" age 30 email "john@example.com"
(integer) 3

> HDEL user:1000 age
(integer) 1

> HDEL user:1000 salary
(integer) 0  # Field didn't exist
```

**Return:**
- Integer number of fields deleted

---

### HGETALL
Returns all fields and values of a hash.

**Syntax:**
```
HGETALL key
```

**Arguments:**
- `key` - The hash key

**Examples:**
```
> HSET user:1000 name "John" age 30
(integer) 2

> HGETALL user:1000
1) "name"
2) "John"
3) "age"
4) "30"

> HGETALL nonexistent
(empty array)
```

**Return:**
- Array of field-value pairs

---

### HKEYS
Returns all field names in a hash.

**Syntax:**
```
HKEYS key
```

**Arguments:**
- `key` - The hash key

**Examples:**
```
> HSET user:1000 name "John" age 30
(integer) 2

> HKEYS user:1000
1) "name"
2) "age"
```

**Return:**
- Array of field names

---

### HVALS
Returns all values in a hash.

**Syntax:**
```
HVALS key
```

**Arguments:**
- `key` - The hash key

**Examples:**
```
> HSET user:1000 name "John" age "30"
(integer) 2

> HVALS user:1000
1) "John"
2) "30"
```

**Return:**
- Array of field values

---

## Key Commands

### DEL
Deletes one or more keys.

**Syntax:**
```
DEL key [key ...]
```

**Arguments:**
- `key` - One or more keys to delete

**Examples:**
```
> SET key1 "value1"
"OK"

> SET key2 "value2"
"OK"

> DEL key1 key2
(integer) 2

> DEL nonexistent
(integer) 0
```

**Return:**
- Integer number of keys deleted

---

### EXISTS
Checks if one or more keys exist.

**Syntax:**
```
EXISTS key [key ...]
```

**Arguments:**
- `key` - One or more keys to check

**Examples:**
```
> SET key1 "value1"
"OK"

> EXISTS key1
(integer) 1

> EXISTS key1 key2 nonexistent
(integer) 1  # Only key1 exists
```

**Return:**
- Integer number of keys that exist

---

### KEYS
Finds all keys matching a pattern.

**Syntax:**
```
KEYS pattern
```

**Arguments:**
- `pattern` - The glob-style pattern
    - `*` matches any number of characters
    - `?` matches single character

**Examples:**
```
> SET user:1 "John"
"OK"

> SET user:2 "Jane"
"OK"

> SET session:abc "data"
"OK"

> KEYS user:*
1) "user:1"
2) "user:2"

> KEYS *
1) "user:1"
2) "user:2"
3) "session:abc"
```

**Return:**
- Array of matching keys

---

### EXPIRE
Sets a key's time to live in seconds.

**Syntax:**
```
EXPIRE key seconds
```

**Arguments:**
- `key` - The key to set expiry for
- `seconds` - Time to live in seconds

**Examples:**
```
> SET mykey "value"
"OK"

> EXPIRE mykey 60
(integer) 1

> EXPIRE nonexistent 60
(integer) 0
```

**Return:**
- `1` if timeout was set
- `0` if key doesn't exist

---

### TTL
Gets the time to live for a key in seconds.

**Syntax:**
```
TTL key
```

**Arguments:**
- `key` - The key to check

**Examples:**
```
> SET mykey "value" EX 60
"OK"

> TTL mykey
(integer) 59

> TTL nonexistent
(integer) -2

> SET permanent "value"
"OK"

> TTL permanent
(integer) -1
```

**Return:**
- TTL in seconds, or:
    - `-2` if key doesn't exist
    - `-1` if key exists but has no expiry

---

## Server Commands

### PING
Tests connection and returns PONG.

**Syntax:**
```
PING
```

**Examples:**
```
> PING
"PONG"
```

**Return:**
- `"PONG"`

---

### ECHO
Echoes the given string.

**Syntax:**
```
ECHO message
```

**Arguments:**
- `message` - The message to echo

**Examples:**
```
> ECHO "Hello World"
"Hello World"
```

**Return:**
- The message string

---

### FLUSHALL
Removes all keys from all databases.

**Syntax:**
```
FLUSHALL
```

**Examples:**
```
> SET key1 "value1"
"OK"

> SET key2 "value2"
"OK"

> FLUSHALL
"OK"

> GET key1
(nil)
```

**Return:**
- `"OK"`

---

### DBSIZE
Returns the number of keys in the database.

**Syntax:**
```
DBSIZE
```

**Examples:**
```
> SET key1 "value1"
"OK"

> SET key2 "value2"
"OK"

> DBSIZE
(integer) 2

> FLUSHALL
"OK"

> DBSIZE
(integer) 0
```

**Return:**
- Integer number of keys

---

## Data Type Summary

| Data Type | Key Commands | Description |
|-----------|--------------|-------------|
| **String** | SET, GET, INCR, DECR | Simple key-value pairs |
| **List** | LPUSH, RPUSH, LPOP, RPOP, LLEN | Ordered collection of strings |
| **Set** | SADD, SREM, SMEMBERS, SISMEMBER | Unordered collection of unique strings |
| **Hash** | HSET, HGET, HDEL, HGETALL, HKEYS, HVALS | Field-value pairs (like objects) |

## Pattern Matching

The `KEYS` command supports simple glob-style patterns:

- `*` - Matches any number of characters
- `?` - Matches exactly one character

**Examples:**
- `KEYS user:*` - All keys starting with "user:"
- `KEYS *:active` - All keys ending with ":active"
- `KEYS ??` - All 2-character keys

## Error Responses

Common error responses:

- `(nil)` - Key doesn't exist or value is null
- `(empty array)` - No elements found
- `ERR wrong number of arguments` - Invalid command syntax
- `ERR value is not an integer` - Type mismatch for numeric operations
- `ERR WRONGTYPE` - Operation on wrong data type

## Quick Reference Card

```
# Strings
SET key value [EX sec]     GET key        INCR key      DECR key

# Lists  
LPUSH key value           RPUSH key value
LPOP key                  RPOP key        LLEN key

# Sets
SADD key member           SREM key member
SMEMBERS key              SISMEMBER key member

# Hashes
HSET key field value      HGET key field
HDEL key field            HGETALL key
HKEYS key                 HVALS key

# Keys
DEL key                   EXISTS key
KEYS pattern              EXPIRE key sec  TTL key

# Server
PING                      ECHO message
FLUSHALL                  DBSIZE
```

This documentation covers all currently implemented commands in your Memora database. The commands are designed to be Redis-compatible for easy migration and familiar usage.