import customtkinter as ctk
import threading
import json
from client_comm import RedisClient, MockClient  # Use our Python client

class RedisClientGUI:
    def __init__(self):
        # Initialize the main window
        ctk.set_appearance_mode("dark")
        ctk.set_default_color_theme("blue")

        self.root = ctk.CTk()
        self.root.title("Memora Redis Client")
        self.root.geometry("800x600")

        # Redis client connection
        self.client = None
        self.is_connected = False

        self.setup_ui()

    def setup_ui(self):
        # Create main frame
        main_frame = ctk.CTkFrame(self.root)
        main_frame.pack(fill="both", expand=True, padx=10, pady=10)

        # Connection frame
        conn_frame = ctk.CTkFrame(main_frame)
        conn_frame.pack(fill="x", padx=10, pady=10)

        ctk.CTkLabel(conn_frame, text="Connection Settings", font=("Arial", 16, "bold")).pack(anchor="w", pady=(0, 10))

        # Host and port inputs
        conn_input_frame = ctk.CTkFrame(conn_frame)
        conn_input_frame.pack(fill="x", pady=5)

        ctk.CTkLabel(conn_input_frame, text="Host:").grid(row=0, column=0, padx=5, pady=5, sticky="w")
        self.host_entry = ctk.CTkEntry(conn_input_frame, placeholder_text="localhost")
        self.host_entry.grid(row=0, column=1, padx=5, pady=5, sticky="ew")
        self.host_entry.insert(0, "localhost")

        ctk.CTkLabel(conn_input_frame, text="Port:").grid(row=0, column=2, padx=5, pady=5, sticky="w")
        self.port_entry = ctk.CTkEntry(conn_input_frame, placeholder_text="6379")
        self.port_entry.grid(row=0, column=3, padx=5, pady=5, sticky="ew")
        self.port_entry.insert(0, "6379")

        conn_input_frame.columnconfigure(1, weight=1)
        conn_input_frame.columnconfigure(3, weight=1)

        # Connection buttons
        conn_button_frame = ctk.CTkFrame(conn_frame)
        conn_button_frame.pack(fill="x", pady=5)

        self.connect_btn = ctk.CTkButton(conn_button_frame, text="Connect", command=self.connect_redis)
        self.connect_btn.pack(side="left", padx=5)

        self.disconnect_btn = ctk.CTkButton(conn_button_frame, text="Disconnect", command=self.disconnect_redis, state="disabled")
        self.disconnect_btn.pack(side="left", padx=5)

        # Mock client checkbox
        self.use_mock_var = ctk.BooleanVar()
        self.mock_checkbox = ctk.CTkCheckBox(conn_button_frame, text="Use Mock Client", variable=self.use_mock_var)
        self.mock_checkbox.pack(side="left", padx=20)

        self.status_label = ctk.CTkLabel(conn_frame, text="Status: Disconnected", text_color="red")
        self.status_label.pack(anchor="w", pady=5)

        # Command input frame
        command_frame = ctk.CTkFrame(main_frame)
        command_frame.pack(fill="x", padx=10, pady=10)

        ctk.CTkLabel(command_frame, text="Redis Command", font=("Arial", 14, "bold")).pack(anchor="w", pady=(0, 5))

        # Command input with examples
        self.command_entry = ctk.CTkEntry(command_frame, placeholder_text="e.g., SET mykey myvalue, GET mykey, KEYS *")
        self.command_entry.pack(fill="x", pady=5)
        self.command_entry.bind("<Return>", lambda e: self.send_command())

        # Quick command buttons
        quick_commands_frame = ctk.CTkFrame(command_frame)
        quick_commands_frame.pack(fill="x", pady=5)

        quick_commands = [
            ("PING", "PING"),
            ("DBSIZE", "DBSIZE"),
            ("FLUSHALL", "FLUSHALL"),
            ("KEYS *", "KEYS *")
        ]

        for text, cmd in quick_commands:
            btn = ctk.CTkButton(quick_commands_frame, text=text, width=100,
                                command=lambda c=cmd: self.quick_command(c))
            btn.pack(side="left", padx=2)

        # Send button
        self.send_btn = ctk.CTkButton(command_frame, text="Send Command", command=self.send_command, state="disabled")
        self.send_btn.pack(anchor="e", pady=5)

        # Results area
        results_frame = ctk.CTkFrame(main_frame)
        results_frame.pack(fill="both", expand=True, padx=10, pady=10)

        ctk.CTkLabel(results_frame, text="Results", font=("Arial", 14, "bold")).pack(anchor="w", pady=(0, 5))

        # Results text box
        self.results_text = ctk.CTkTextbox(results_frame, wrap="word")
        self.results_text.pack(fill="both", expand=True, pady=5)

        # Key-Value operations frame
        kv_frame = ctk.CTkFrame(main_frame)
        kv_frame.pack(fill="x", padx=10, pady=10)

        ctk.CTkLabel(kv_frame, text="Key-Value Operations", font=("Arial", 14, "bold")).pack(anchor="w", pady=(0, 5))

        # Key-Value input fields
        kv_input_frame = ctk.CTkFrame(kv_frame)
        kv_input_frame.pack(fill="x", pady=5)

        ctk.CTkLabel(kv_input_frame, text="Key:").grid(row=0, column=0, padx=5, pady=5, sticky="w")
        self.key_entry = ctk.CTkEntry(kv_input_frame)
        self.key_entry.grid(row=0, column=1, padx=5, pady=5, sticky="ew")

        ctk.CTkLabel(kv_input_frame, text="Value:").grid(row=0, column=2, padx=5, pady=5, sticky="w")
        self.value_entry = ctk.CTkEntry(kv_input_frame)
        self.value_entry.grid(row=0, column=3, padx=5, pady=5, sticky="ew")

        kv_input_frame.columnconfigure(1, weight=1)
        kv_input_frame.columnconfigure(3, weight=1)

        # KV operation buttons
        kv_buttons_frame = ctk.CTkFrame(kv_frame)
        kv_buttons_frame.pack(fill="x", pady=5)

        self.set_btn = ctk.CTkButton(kv_buttons_frame, text="SET", command=self.set_key, state="disabled")
        self.set_btn.pack(side="left", padx=2)

        self.get_btn = ctk.CTkButton(kv_buttons_frame, text="GET", command=self.get_key, state="disabled")
        self.get_btn.pack(side="left", padx=2)

        self.del_btn = ctk.CTkButton(kv_buttons_frame, text="DEL", command=self.del_key, state="disabled")
        self.del_btn.pack(side="left", padx=2)

        self.exists_btn = ctk.CTkButton(kv_buttons_frame, text="EXISTS", command=self.exists_key, state="disabled")
        self.exists_btn.pack(side="left", padx=2)

    def connect_redis(self):
        def connect_thread():
            try:
                host = self.host_entry.get()
                port = self.port_entry.get()

                if self.use_mock_var.get():
                    # Use mock client
                    self.client = MockClient(host, port)
                    self.is_connected = True
                    self.root.after(0, self.update_connection_status, True, "Connected to Mock Client")
                else:
                    # Use real client
                    self.client = RedisClient(host, int(port))
                    self.is_connected = True
                    self.root.after(0, self.update_connection_status, True, "Connected successfully")

            except Exception as e:
                self.root.after(0, self.update_connection_status, False, f"Connection failed: {str(e)}")

        # Run connection in separate thread
        threading.Thread(target=connect_thread, daemon=True).start()
        self.status_label.configure(text="Connecting...", text_color="yellow")

    def disconnect_redis(self):
        if self.client:
            self.client.close()
            self.client = None
            self.is_connected = False
            self.update_connection_status(False, "Disconnected")

    def update_connection_status(self, connected, message):
        self.is_connected = connected

        if connected:
            self.status_label.configure(text=f"Status: {message}", text_color="green")
            self.connect_btn.configure(state="disabled")
            self.disconnect_btn.configure(state="normal")
            self.send_btn.configure(state="normal")
            self.set_btn.configure(state="normal")
            self.get_btn.configure(state="normal")
            self.del_btn.configure(state="normal")
            self.exists_btn.configure(state="normal")
        else:
            self.status_label.configure(text=f"Status: {message}", text_color="red")
            self.connect_btn.configure(state="normal")
            self.disconnect_btn.configure(state="disabled")
            self.send_btn.configure(state="disabled")
            self.set_btn.configure(state="disabled")
            self.get_btn.configure(state="disabled")
            self.del_btn.configure(state="disabled")
            self.exists_btn.configure(state="disabled")

        self.log_result(f"=== {message} ===\n")

    def send_command(self):
        if not self.is_connected or not self.client:
            self.log_result("Error: Not connected to Redis server\n")
            return

        command_str = self.command_entry.get().strip()
        if not command_str:
            return

        def command_thread():
            try:
                # Parse command
                parts = command_str.split()
                result = self.client.send_command(parts)

                # Format result
                formatted_result = self.format_result(result)
                self.root.after(0, self.log_result, f"> {command_str}\n{formatted_result}\n\n")

            except Exception as e:
                self.root.after(0, self.log_result, f"> {command_str}\nError: {str(e)}\n\n")

        threading.Thread(target=command_thread, daemon=True).start()

    def quick_command(self, command):
        self.command_entry.delete(0, "end")
        self.command_entry.insert(0, command)
        self.send_command()

    def set_key(self):
        key = self.key_entry.get().strip()
        value = self.value_entry.get().strip()

        if not key:
            self.log_result("Error: Key cannot be empty\n")
            return

        self.command_entry.delete(0, "end")
        self.command_entry.insert(0, f"SET {key} {value}")
        self.send_command()

    def get_key(self):
        key = self.key_entry.get().strip()

        if not key:
            self.log_result("Error: Key cannot be empty\n")
            return

        self.command_entry.delete(0, "end")
        self.command_entry.insert(0, f"GET {key}")
        self.send_command()

    def del_key(self):
        key = self.key_entry.get().strip()

        if not key:
            self.log_result("Error: Key cannot be empty\n")
            return

        self.command_entry.delete(0, "end")
        self.command_entry.insert(0, f"DEL {key}")
        self.send_command()

    def exists_key(self):
        key = self.key_entry.get().strip()

        if not key:
            self.log_result("Error: Key cannot be empty\n")
            return

        self.command_entry.delete(0, "end")
        self.command_entry.insert(0, f"EXISTS {key}")
        self.send_command()

    def format_result(self, result):
        if result is None:
            return "(nil)"
        elif isinstance(result, str):
            return f'"{result}"'
        elif isinstance(result, int):
            return f"(integer) {result}"
        elif isinstance(result, list):
            if len(result) == 0:
                return "(empty list or set)"
            else:
                formatted = []
                for i, item in enumerate(result):
                    formatted.append(f"{i+1}) {self.format_result(item)}")
                return "\n".join(formatted)
        elif isinstance(result, dict):
            return json.dumps(result, indent=2)
        else:
            return str(result)

    def log_result(self, text):
        self.results_text.insert("end", text)
        self.results_text.see("end")

    def run(self):
        self.root.mainloop()

if __name__ == "__main__":
    app = RedisClientGUI()
    app.run()