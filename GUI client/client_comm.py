import socket
import json

class RedisClient:
    def __init__(self, host='localhost', port=6379):
        self.host = host
        self.port = port
        self.socket = None
        self.connect()

    def connect(self):
        """Connect to the Redis server"""
        try:
            self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.socket.connect((self.host, self.port))
            return True
        except Exception as e:
            print(f"Connection error: {e}")
            return False

    def close(self):
        """Close the connection"""
        if self.socket:
            self.socket.close()

    def send_command(self, command):
        """Send a command to the Redis server using RESP protocol"""
        if not self.socket:
            if not self.connect():
                raise Exception("Not connected to server")

        # Build RESP array
        resp = f"*{len(command)}\r\n"
        for arg in command:
            resp += f"${len(arg)}\r\n{arg}\r\n"

        # Send command
        self.socket.send(resp.encode())

        # Read response
        return self.read_response()

    def read_response(self):
        """Read and parse RESP response"""
        line = self.socket.makefile().readline().strip()

        if not line:
            raise Exception("Empty response")

        resp_type = line[0]
        content = line[1:]

        if resp_type == '+':  # Simple string
            return content
        elif resp_type == '-':  # Error
            raise Exception(content)
        elif resp_type == ':':  # Integer
            return int(content)
        elif resp_type == '$':  # Bulk string
            length = int(content)
            if length == -1:
                return None
            data = self.socket.recv(length)
            # Read the trailing \r\n
            self.socket.recv(2)
            return data.decode()
        elif resp_type == '*':  # Array
            length = int(content)
            if length == -1:
                return None
            array = []
            for _ in range(length):
                array.append(self.read_response())
            return array
        else:
            raise Exception(f"Unknown response type: {resp_type}")

    # Convenience methods
    def set(self, key, value):
        return self.send_command(['SET', key, value])

    def get(self, key):
        return self.send_command(['GET', key])

    def delete(self, *keys):
        return self.send_command(['DEL'] + list(keys))

    def exists(self, *keys):
        return self.send_command(['EXISTS'] + list(keys))

    def keys(self, pattern):
        return self.send_command(['KEYS', pattern])

    def ping(self):
        return self.send_command(['PING'])

    def flushall(self):
        return self.send_command(['FLUSHALL'])

    def dbsize(self):
        return self.send_command(['DBSIZE'])

    def echo(self, message):
        return self.send_command(['ECHO', message])

# Mock client for testing when server is not available
class MockClient:
    def __init__(self, host='localhost', port=6379):
        self.data = {}
        print(f"Mock client connected to {host}:{port}")

    def close(self):
        print("Mock client disconnected")

    def send_command(self, command):
        cmd = command[0].upper()

        if cmd == 'PING':
            return "PONG"
        elif cmd == 'SET' and len(command) >= 3:
            self.data[command[1]] = command[2]
            return "OK"
        elif cmd == 'GET' and len(command) >= 2:
            return self.data.get(command[1], None)
        elif cmd == 'DEL' and len(command) >= 2:
            count = 0
            for key in command[1:]:
                if key in self.data:
                    del self.data[key]
                    count += 1
            return count
        elif cmd == 'EXISTS' and len(command) >= 2:
            count = 0
            for key in command[1:]:
                if key in self.data:
                    count += 1
            return count
        elif cmd == 'KEYS' and len(command) >= 2:
            pattern = command[1]
            if pattern == '*':
                return list(self.data.keys())
            return []
        elif cmd == 'FLUSHALL':
            self.data.clear()
            return "OK"
        elif cmd == 'DBSIZE':
            return len(self.data)
        elif cmd == 'ECHO' and len(command) >= 2:
            return command[1]
        else:
            return f"Unknown command: {cmd}"

    # Convenience methods to match RedisClient interface
    def set(self, key, value):
        return self.send_command(['SET', key, value])

    def get(self, key):
        return self.send_command(['GET', key])

    def delete(self, *keys):
        return self.send_command(['DEL'] + list(keys))

    def exists(self, *keys):
        return self.send_command(['EXISTS'] + list(keys))

    def keys(self, pattern):
        return self.send_command(['KEYS', pattern])

    def ping(self):
        return self.send_command(['PING'])

    def flushall(self):
        return self.send_command(['FLUSHALL'])

    def dbsize(self):
        return self.send_command(['DBSIZE'])

    def echo(self, message):
        return self.send_command(['ECHO', message])