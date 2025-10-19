package commands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"Memora/store"
)

type CommandHandler struct {
	store *store.DataStore
}

func NewCommandHandler(store *store.DataStore) *CommandHandler {
	return &CommandHandler{store: store}
}

func (h *CommandHandler) HandleCommand(command []string) interface{} {
	if len(command) == 0 {
		return nil
	}

	cmd := strings.ToUpper(command[0])
	args := command[1:]

	switch cmd {
	// String commands
	case "SET":
		return h.handleSet(args)
	case "GET":
		return h.handleGet(args)
	case "DEL":
		return h.handleDel(args)
	case "EXISTS":
		return h.handleExists(args)
	case "KEYS":
		return h.handleKeys(args)
	case "TTL":
		return h.handleTTL(args)
	case "EXPIRE":
		return h.handleExpire(args)
	case "INCR":
		return h.handleIncr(args)
	case "DECR":
		return h.handleDecr(args)

	// List commands
	case "LPUSH":
		return h.handleLPush(args)
	case "RPUSH":
		return h.handleRPush(args)
	case "LPOP":
		return h.handleLPop(args)
	case "RPOP":
		return h.handleRPop(args)
	case "LLEN":
		return h.handleLLen(args)

	// Set commands
	case "SADD":
		return h.handleSAdd(args)
	case "SREM":
		return h.handleSRem(args)
	case "SMEMBERS":
		return h.handleSMembers(args)
	case "SISMEMBER":
		return h.handleSIsMember(args)

	// Hash commands
	case "HSET":
		return h.handleHSet(args)
	case "HGET":
		return h.handleHGet(args)
	case "HDEL":
		return h.handleHDel(args)
	case "HGETALL":
		return h.handleHGetAll(args)
	case "HKEYS":
		return h.handleHKeys(args)
	case "HVALS":
		return h.handleHVals(args)

	// Server commands
	case "PING":
		return "PONG"
	case "ECHO":
		return h.handleEcho(args)
	case "FLUSHALL":
		return h.handleFlushAll(args)
	case "DBSIZE":
		return h.handleDBSize(args)
	case "COMMAND":
		return "OK" // Basic command support

	// Sort commands
	case "ZRANGEBYLEX":
		return h.handleZRANGEBYLEX(args)

	default:
		// If it's not a recognized command, treat it as GET
		// This handles cases where user types just the key name
		if len(command) == 1 {
			return h.handleGet(command)
		}
		return nil
	}
}

// String command handlers
func (h *CommandHandler) handleSet(args []string) interface{} {
	if len(args) < 2 {
		return "ERR wrong number of arguments for 'set' command"
	}

	key := args[0]
	value := args[1]
	var ttl time.Duration

	// Handle SET key value EX seconds or PX milliseconds
	if len(args) > 3 {
		option := strings.ToUpper(args[2])
		if option == "EX" {
			seconds, err := strconv.Atoi(args[3])
			if err != nil {
				return "ERR invalid expire time in 'set' command"
			}
			ttl = time.Duration(seconds) * time.Second
		} else if option == "PX" {
			millis, err := strconv.Atoi(args[3])
			if err != nil {
				return "ERR invalid expire time in 'set' command"
			}
			ttl = time.Duration(millis) * time.Millisecond
		}
	}

	h.store.Set(key, value, ttl)
	return "OK"
}

func (h *CommandHandler) handleGet(args []string) interface{} {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'get' command"
	}

	value, exists := h.store.Get(args[0])
	if !exists {
		return nil
	}

	// Convert to string if it's stored as string
	if str, ok := value.(string); ok {
		return str
	}

	return fmt.Sprintf("%v", value)
}

func (h *CommandHandler) handleDel(args []string) interface{} {
	if len(args) == 0 {
		return 0
	}

	deleted := 0
	for _, key := range args {
		if h.store.Delete(key) {
			deleted++
		}
	}
	return deleted
}

func (h *CommandHandler) handleExists(args []string) interface{} {
	if len(args) == 0 {
		return 0
	}

	count := 0
	for _, key := range args {
		if h.store.Exists(key) {
			count++
		}
	}
	return count
}

func (h *CommandHandler) handleKeys(args []string) interface{} {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'keys' command"
	}

	pattern := args[0]
	if pattern == "" {
		pattern = "*"
	}

	keys := h.store.Keys(pattern)
	result := make([]interface{}, len(keys))
	for i, key := range keys {
		result[i] = key
	}
	return result
}

func (h *CommandHandler) handleTTL(args []string) interface{} {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'ttl' command"
	}

	return h.store.TTL(args[0])
}

func (h *CommandHandler) handleExpire(args []string) interface{} {
	if len(args) != 2 {
		return "ERR wrong number of arguments for 'expire' command"
	}

	key := args[0]
	seconds, err := strconv.Atoi(args[1])
	if err != nil {
		return 0
	}

	ttl := time.Duration(seconds) * time.Second
	if h.store.Expire(key, ttl) {
		return 1
	}
	return 0
}

func (h *CommandHandler) handleIncr(args []string) interface{} {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'incr' command"
	}

	key := args[0]
	value, exists := h.store.Get(key)
	if !exists {
		h.store.Set(key, "1", 0)
		return 1
	}

	strValue, ok := value.(string)
	if !ok {
		return "ERR value is not an integer or out of range"
	}

	num, err := strconv.Atoi(strValue)
	if err != nil {
		return "ERR value is not an integer or out of range"
	}

	num++
	h.store.Set(key, strconv.Itoa(num), 0)
	return num
}

func (h *CommandHandler) handleDecr(args []string) interface{} {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'decr' command"
	}

	key := args[0]
	value, exists := h.store.Get(key)
	if !exists {
		h.store.Set(key, "-1", 0)
		return -1
	}

	strValue, ok := value.(string)
	if !ok {
		return "ERR value is not an integer or out of range"
	}

	num, err := strconv.Atoi(strValue)
	if err != nil {
		return "ERR value is not an integer or out of range"
	}

	num--
	h.store.Set(key, strconv.Itoa(num), 0)
	return num
}

// List command handlers
func (h *CommandHandler) handleLPush(args []string) interface{} {
	if len(args) < 2 {
		return "ERR wrong number of arguments for 'lpush' command"
	}

	key := args[0]
	values := make([]interface{}, len(args)-1)
	for i, arg := range args[1:] {
		values[i] = arg
	}

	return h.store.LPush(key, values...)
}

func (h *CommandHandler) handleRPush(args []string) interface{} {
	if len(args) < 2 {
		return "ERR wrong number of arguments for 'rpush' command"
	}

	key := args[0]
	values := make([]interface{}, len(args)-1)
	for i, arg := range args[1:] {
		values[i] = arg
	}

	return h.store.RPush(key, values...)
}

func (h *CommandHandler) handleLPop(args []string) interface{} {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'lpop' command"
	}

	return h.store.LPop(args[0])
}

func (h *CommandHandler) handleRPop(args []string) interface{} {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'rpop' command"
	}

	return h.store.RPop(args[0])
}

func (h *CommandHandler) handleLLen(args []string) interface{} {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'llen' command"
	}

	return h.store.LLen(args[0])
}

// Set command handlers
func (h *CommandHandler) handleSAdd(args []string) interface{} {
	if len(args) < 2 {
		return "ERR wrong number of arguments for 'sadd' command"
	}

	key := args[0]
	members := make([]interface{}, len(args)-1)
	for i, arg := range args[1:] {
		members[i] = arg
	}

	return h.store.SAdd(key, members...)
}

func (h *CommandHandler) handleSRem(args []string) interface{} {
	if len(args) < 2 {
		return "ERR wrong number of arguments for 'srem' command"
	}

	key := args[0]
	members := make([]interface{}, len(args)-1)
	for i, arg := range args[1:] {
		members[i] = arg
	}

	return h.store.SRem(key, members...)
}

func (h *CommandHandler) handleSMembers(args []string) interface{} {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'smembers' command"
	}

	members := h.store.SMembers(args[0])
	if members == nil {
		return []interface{}{}
	}
	return members
}

func (h *CommandHandler) handleSIsMember(args []string) interface{} {
	if len(args) != 2 {
		return "ERR wrong number of arguments for 'sismember' command"
	}

	if h.store.SIsMember(args[0], args[1]) {
		return 1
	}
	return 0
}

// Hash command handlers
func (h *CommandHandler) handleHSet(args []string) interface{} {
	if len(args) < 3 || len(args)%2 != 1 {
		return "ERR wrong number of arguments for 'hset' command"
	}

	key := args[0]
	added := 0

	for i := 1; i < len(args); i += 2 {
		field := args[i]
		value := args[i+1]
		if h.store.HSet(key, field, value) {
			added++
		}
	}

	return added
}

func (h *CommandHandler) handleHGet(args []string) interface{} {
	if len(args) != 2 {
		return "ERR wrong number of arguments for 'hget' command"
	}

	return h.store.HGet(args[0], args[1])
}

func (h *CommandHandler) handleHDel(args []string) interface{} {
	if len(args) < 2 {
		return "ERR wrong number of arguments for 'hdel' command"
	}

	return h.store.HDel(args[0], args[1:]...)
}

func (h *CommandHandler) handleHGetAll(args []string) interface{} {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'hgetall' command"
	}

	hash := h.store.HGetAll(args[0])
	if hash == nil {
		return []interface{}{}
	}

	result := make([]interface{}, 0, len(hash)*2)
	for field, value := range hash {
		result = append(result, field, value)
	}
	return result
}

func (h *CommandHandler) handleHKeys(args []string) interface{} {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'hkeys' command"
	}

	hash := h.store.HGetAll(args[0])
	if hash == nil {
		return []interface{}{}
	}

	keys := make([]interface{}, 0, len(hash))
	for field := range hash {
		keys = append(keys, field)
	}
	return keys
}

func (h *CommandHandler) handleHVals(args []string) interface{} {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'hvals' command"
	}

	hash := h.store.HGetAll(args[0])
	if hash == nil {
		return []interface{}{}
	}

	values := make([]interface{}, 0, len(hash))
	for _, value := range hash {
		values = append(values, value)
	}
	return values
}

// Server command handlers
func (h *CommandHandler) handleEcho(args []string) interface{} {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'echo' command"
	}
	return args[0]
}

func (h *CommandHandler) handleFlushAll(args []string) interface{} {
	h.store.FlushAll()
	return "OK"
}

func (h *CommandHandler) handleDBSize(args []string) interface{} {
	keys := h.store.Keys("*")
	return len(keys)
}

func (h *CommandHandler) handleZRANGEBYLEX(args []string) interface{} {
	if len(args) != 2 {
		return "ERR wrong number of arguments for 'ZRANGEBYLEX' command"
	}

	key := args[0]
	order := args[1] // "I" for increasing, "D" for decreasing

	// Get members (assumed to return []interface{})
	membersIface := h.store.SMembers(key)
	if membersIface == nil {
		return []interface{}{}
	}

	// Convert []interface{} to []string
	members := make([]string, 0, len(membersIface))
	for _, v := range membersIface {
		str, ok := v.(string)
		if ok {
			members = append(members, str)
		}
	}

	// Sort members
	switch order {
	case "I":
		sort.Strings(members)
	case "D":
		sort.Sort(sort.Reverse(sort.StringSlice(members)))
	default:
		return "ERR invalid sort order; use 'I' or 'D'"
	}

	// Convert back to []interface{}
	result := make([]interface{}, len(members))
	for i, v := range members {
		result[i] = v
	}

	return result
}
