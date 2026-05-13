package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"adb-broadcast-sender/internal/app"
	"adb-broadcast-sender/internal/models"
)

// Server holds the HTTP server state
type Server struct {
	state   *app.AppState
	mu      sync.RWMutex
	clients map[chan string]struct{}
}

// API Response types
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type DeviceInfo struct {
	Serial     string `json:"serial"`
	Model      string `json:"model"`
	State      string `json:"state"`
	Connection string `json:"connection"`
	IP         string `json:"ip,omitempty"`
	Port       int    `json:"port,omitempty"`
}

type BroadcastInfo struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	DeviceID  string `json:"device_id"`
	Timestamp string `json:"timestamp"`
	Status    string `json:"status"`
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
}

func main() {
	// Initialize application state
	log.Println("Initializing application state...")
	state, err := app.NewAppState()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	log.Println("Application state initialized successfully")

	// Create server
	server := &Server{
		state:   state,
		clients: make(map[chan string]struct{}),
	}

	// Start auto refresh
	ctx := context.Background()
	state.StartAutoRefresh(ctx)

	// Set up HTTP routes
	http.HandleFunc("/", server.handleIndex)
	http.HandleFunc("/api/devices", server.handleDevices)
	http.HandleFunc("/api/select-device", server.handleSelectDevice)
	http.HandleFunc("/api/connect", server.handleConnect)
	http.HandleFunc("/api/disconnect", server.handleDisconnect)
	http.HandleFunc("/api/send", server.handleSend)
	http.HandleFunc("/api/history", server.handleHistory)
	http.HandleFunc("/api/clear-history", server.handleClearHistory)
	http.HandleFunc("/api/events", server.handleEvents)

	// Start server
	port := 8080
	url := fmt.Sprintf("http://localhost:%d", port)
	log.Printf("Starting server on %s", url)

	// Open browser
	go func() {
		time.Sleep(500 * time.Millisecond)
		openBrowser(url)
	}()

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// handleIndex serves the main HTML page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("index").Parse(indexHTML)
	if err != nil {
		log.Printf("Template parse error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
	}{
		Title: "ADB Broadcast Sender",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execute error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleDevices returns the list of connected devices
func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	devices := s.state.GetConnectedDevices()

	var result []DeviceInfo
	for _, d := range devices {
		result = append(result, DeviceInfo{
			Serial:     d.Serial,
			Model:      d.Model,
			State:      d.State,
			Connection: d.Connection,
			IP:         d.IPAddress,
			Port:       d.Port,
		})
	}

	// Return empty array instead of null
	if result == nil {
		result = []DeviceInfo{}
	}

	json.NewEncoder(w).Encode(Response{
		Success: true,
		Data:    result,
	})
}

// handleSelectDevice selects a device
func (s *Server) handleSelectDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		json.NewEncoder(w).Encode(Response{Success: false, Message: "Method not allowed"})
		return
	}

	var req struct {
		Serial string `json:"serial"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(Response{Success: false, Message: "Invalid request"})
		return
	}

	device := s.state.DeviceManager.GetDevice(req.Serial)
	if device == nil {
		json.NewEncoder(w).Encode(Response{Success: false, Message: "Device not found"})
		return
	}

	s.state.SelectDevice(device)
	s.notifyClients("device-selected")

	json.NewEncoder(w).Encode(Response{
		Success: true,
		Message: fmt.Sprintf("Selected device: %s", device.GetDisplayName()),
	})
}

// handleConnect connects to a device via WiFi
func (s *Server) handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		json.NewEncoder(w).Encode(Response{Success: false, Message: "Method not allowed"})
		return
	}

	var req struct {
		IP   string `json:"ip"`
		Port int    `json:"port"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(Response{Success: false, Message: "Invalid request"})
		return
	}

	if req.IP == "" {
		json.NewEncoder(w).Encode(Response{Success: false, Message: "IP address is required"})
		return
	}

	if req.Port == 0 {
		req.Port = 5555 // Default ADB port
	}

	err := s.state.DeviceManager.ConnectWiFi(req.IP, req.Port)
	if err != nil {
		json.NewEncoder(w).Encode(Response{Success: false, Message: err.Error()})
		return
	}

	s.notifyClients("device-connected")
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Message: fmt.Sprintf("Connected to %s:%d", req.IP, req.Port),
	})
}

// handleDisconnect disconnects from a WiFi device
func (s *Server) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		json.NewEncoder(w).Encode(Response{Success: false, Message: "Method not allowed"})
		return
	}

	var req struct {
		IP   string `json:"ip"`
		Port int    `json:"port"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(Response{Success: false, Message: "Invalid request"})
		return
	}

	if req.Port == 0 {
		req.Port = 5555
	}

	err := s.state.DeviceManager.DisconnectWiFi(req.IP, req.Port)
	if err != nil {
		json.NewEncoder(w).Encode(Response{Success: false, Message: err.Error()})
		return
	}

	s.notifyClients("device-disconnected")
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Message: fmt.Sprintf("Disconnected from %s:%d", req.IP, req.Port),
	})
}

// handleSend sends a broadcast
func (s *Server) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		json.NewEncoder(w).Encode(Response{Success: false, Message: "Method not allowed"})
		return
	}

	var req struct {
		Content string `json:"content"`
		Action  string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(Response{Success: false, Message: "Invalid request"})
		return
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		json.NewEncoder(w).Encode(Response{Success: false, Message: "Content is required"})
		return
	}

	device := s.state.GetSelectedDevice()
	if device == nil {
		json.NewEncoder(w).Encode(Response{Success: false, Message: "No device selected"})
		return
	}

	var result *models.Broadcast
	var err error

	action := strings.TrimSpace(req.Action)
	if action != "" {
		result, err = s.state.SendBroadcastWithAction(action, content)
	} else {
		result, err = s.state.SendBroadcast(content)
	}

	if err != nil {
		json.NewEncoder(w).Encode(Response{Success: false, Message: err.Error()})
		return
	}

	s.notifyClients("broadcast-sent")

	json.NewEncoder(w).Encode(Response{
		Success: result.IsSuccess(),
		Message: result.GetStatus(),
		Data: BroadcastInfo{
			ID:        result.ID,
			Content:   result.Content,
			DeviceID:  result.DeviceID,
			Timestamp: result.GetFormattedTimestamp(),
			Status:    result.GetStatus(),
			Output:    result.Output,
			Error:     result.Error,
		},
	})
}

// handleHistory returns the broadcast history
func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	history := s.state.GetBroadcastHistory()

	var result []BroadcastInfo
	for _, b := range history {
		result = append(result, BroadcastInfo{
			ID:        b.ID,
			Content:   b.Content,
			DeviceID:  b.DeviceID,
			Timestamp: b.GetFormattedTimestamp(),
			Status:    b.GetStatus(),
			Output:    b.Output,
			Error:     b.Error,
		})
	}

	if result == nil {
		result = []BroadcastInfo{}
	}

	json.NewEncoder(w).Encode(Response{
		Success: true,
		Data:    result,
	})
}

// handleClearHistory clears the broadcast history
func (s *Server) handleClearHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		json.NewEncoder(w).Encode(Response{Success: false, Message: "Method not allowed"})
		return
	}

	s.state.ClearHistory()
	s.notifyClients("history-cleared")

	json.NewEncoder(w).Encode(Response{Success: true, Message: "History cleared"})
}

// handleEvents handles Server-Sent Events for real-time updates
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create client channel
	ch := make(chan string, 10)
	s.mu.Lock()
	s.clients[ch] = struct{}{}
	s.mu.Unlock()

	// Clean up on disconnect
	defer func() {
		s.mu.Lock()
		delete(s.clients, ch)
		s.mu.Unlock()
		close(ch)
	}()

	// Send initial message
	fmt.Fprintf(w, "data: connected\n\n")
	w.(http.Flusher).Flush()

	// Stream events
	for {
		select {
		case msg := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// notifyClients sends an event to all connected clients
func (s *Server) notifyClients(event string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for ch := range s.clients {
		select {
		case ch <- event:
		default:
		}
	}
}

// openBrowser opens the default browser
func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Printf("Failed to open browser: %v", err)
	}
}

// HTML template
const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #f5f5f5;
            color: #333;
            line-height: 1.6;
        }
        .container { max-width: 900px; margin: 0 auto; padding: 20px; }
        h1 { text-align: center; margin-bottom: 30px; color: #2196F3; }
        .card {
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            padding: 20px;
            margin-bottom: 20px;
        }
        .card h2 { margin-bottom: 15px; color: #666; font-size: 16px; }
        .form-group { margin-bottom: 15px; }
        .form-row { display: flex; gap: 10px; }
        .form-row .form-group { flex: 1; }
        label { display: block; margin-bottom: 5px; font-weight: 500; color: #555; }
        input, select, textarea {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 14px;
        }
        textarea { min-height: 100px; resize: vertical; }
        .btn {
            display: inline-block;
            padding: 10px 20px;
            background: #2196F3;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
            margin-right: 10px;
            margin-top: 5px;
        }
        .btn:hover { background: #1976D2; }
        .btn-success { background: #4CAF50; }
        .btn-success:hover { background: #388E3C; }
        .btn-secondary { background: #757575; }
        .btn-secondary:hover { background: #616161; }
        .btn-danger { background: #f44336; }
        .btn-danger:hover { background: #d32f2f; }
        .btn-quick {
            display: inline-block;
            padding: 6px 12px;
            background: #E3F2FD;
            color: #1976D2;
            border: 1px solid #90CAF9;
            border-radius: 4px;
            cursor: pointer;
            font-size: 13px;
            transition: all 0.2s;
        }
        .btn-quick:hover {
            background: #1976D2;
            color: white;
            border-color: #1976D2;
        }
        .status { padding: 10px; border-radius: 4px; margin-top: 10px; }
        .status.success { background: #E8F5E9; color: #2E7D32; }
        .status.error { background: #FFEBEE; color: #C62828; }
        .status.info { background: #E3F2FD; color: #1565C0; }
        .device-info { padding: 10px; background: #f9f9f9; border-radius: 4px; margin-top: 10px; }
        .history-item { padding: 15px; border-bottom: 1px solid #eee; }
        .history-item:last-child { border-bottom: none; }
        .history-item.success { border-left: 3px solid #4CAF50; }
        .history-item.failed { border-left: 3px solid #f44336; }
        .history-item .timestamp { color: #999; font-size: 12px; }
        .history-item .content { margin: 5px 0; word-break: break-all; }
        .empty-state { text-align: center; padding: 40px; color: #999; }
        .wifi-section { background: #f0f7ff; padding: 15px; border-radius: 4px; margin-top: 10px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>ADB Broadcast Sender</h1>

        <div class="card">
            <h2>Device Selection</h2>
            <div class="form-group">
                <label for="device">Select Device:</label>
                <select id="device">
                    <option value="">-- Loading devices... --</option>
                </select>
            </div>
            <button class="btn btn-secondary" onclick="refreshDevices()">Refresh</button>
            <div id="deviceInfo" class="device-info" style="display:none;"></div>

            <div class="wifi-section">
                <h3 style="margin-bottom:10px; font-size:14px; color:#1976D2;">WiFi Connect</h3>
                <div class="form-row">
                    <div class="form-group">
                        <label for="wifiIP">IP Address:</label>
                        <input type="text" id="wifiIP" placeholder="192.168.1.100">
                    </div>
                    <div class="form-group" style="flex: 0 0 120px;">
                        <label for="wifiPort">Port:</label>
                        <input type="text" id="wifiPort" value="5555" placeholder="5555">
                    </div>
                </div>
                <button class="btn btn-success" onclick="connectWiFi()">Connect</button>
                <button class="btn btn-danger" onclick="disconnectWiFi()">Disconnect</button>
            </div>
        </div>

        <div class="card">
            <h2>Broadcast Settings</h2>
            <div class="form-group">
                <label for="action">Action (default: com.sunmi.scanner.ACTION_DATA_CODE_RECEIVED):</label>
                <input type="text" id="action" placeholder="com.sunmi.scanner.ACTION_DATA_CODE_RECEIVED">
            </div>
            <div class="form-group">
                <label for="content">Content:</label>
                <textarea id="content" placeholder="Enter broadcast content..."></textarea>
            </div>

            <div class="quick-content-section" style="margin-bottom: 15px;">
                <h3 style="margin-bottom: 10px; font-size: 14px; color: #1976D2;">Quick Content</h3>
                <div class="quick-content-buttons" style="display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 10px;">
                    <button class="btn-quick" onclick="setQuickContent('Test QR Code')">Test QR Code</button>
                    <button class="btn-quick" onclick="setQuickContent('1234567890')">Test Numbers</button>
                    <button class="btn-quick" onclick="setQuickContent('ABCDEFG')">Test Letters</button>
                    <button class="btn-quick" onclick="setQuickContent('https://example.com')">Test URL</button>
                    <button class="btn-quick" onclick="setQuickContent('{&quot;code&quot;:&quot;12345&quot;,&quot;type&quot;:&quot;qr&quot;}')">Test JSON</button>
                </div>
                <div class="custom-quick-content" style="display: flex; gap: 8px;">
                    <input type="text" id="customContent" placeholder="Add custom content..." style="flex: 1; padding: 8px; border: 1px solid #ddd; border-radius: 4px;">
                    <button class="btn btn-secondary" onclick="addCustomContent()" style="padding: 8px 12px;">Add</button>
                </div>
            </div>

            <button class="btn" onclick="sendBroadcast()">Send Broadcast</button>
            <button class="btn btn-secondary" onclick="clearHistory()">Clear History</button>
        </div>

        <div class="card">
            <h2>Broadcast History</h2>
            <div id="history"></div>
        </div>

        <div id="status" class="status info" style="display:none;"></div>
    </div>

    <script>
        var selectedDevice = null;

        document.addEventListener('DOMContentLoaded', function() {
            refreshDevices();
            loadHistory();
            setupEventSource();
        });

        function setupEventSource() {
            var eventSource = new EventSource('/api/events');
            eventSource.onmessage = function(e) {
                if (e.data === 'device-selected' || e.data === 'device-connected' ||
                    e.data === 'device-disconnected' || e.data === 'broadcast-sent' ||
                    e.data === 'history-cleared') {
                    refreshDevices();
                    loadHistory();
                }
            };
        }

        async function refreshDevices() {
            try {
                var res = await fetch('/api/devices');
                var data = await res.json();
                var select = document.getElementById('device');
                select.innerHTML = '<option value="">-- Select a device --</option>';

                if (data.data && data.data.length > 0) {
                    for (var i = 0; i < data.data.length; i++) {
                        var device = data.data[i];
                        var option = document.createElement('option');
                        option.value = device.serial;
                        option.textContent = device.model ? device.model + ' (' + device.serial + ')' : device.serial;
                        option.dataset.model = device.model || '';
                        option.dataset.state = device.state;
                        option.dataset.connection = device.connection;
                        option.dataset.ip = device.ip || '';
                        option.dataset.port = device.port || 5555;
                        select.appendChild(option);
                    }
                    showStatus('Found ' + data.data.length + ' device(s)', 'success');
                } else {
                    select.innerHTML = '<option value="">-- No devices found --</option>';
                    showStatus('No devices found. Connect via WiFi or check USB connection.', 'info');
                }
            } catch (err) {
                showStatus('Failed to refresh devices: ' + err.message, 'error');
            }
        }

        document.getElementById('device').addEventListener('change', async function(e) {
            var serial = e.target.value;
            if (!serial) {
                selectedDevice = null;
                document.getElementById('deviceInfo').style.display = 'none';
                return;
            }

            var option = e.target.selectedOptions[0];
            selectedDevice = {
                serial: serial,
                model: option.dataset.model,
                state: option.dataset.state,
                connection: option.dataset.connection,
                ip: option.dataset.ip,
                port: option.dataset.port
            };

            try {
                await fetch('/api/select-device', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ serial: serial })
                });

                var infoDiv = document.getElementById('deviceInfo');
                infoDiv.innerHTML = '<strong>' + (selectedDevice.model || selectedDevice.serial) + '</strong><br>' +
                    'Serial: ' + selectedDevice.serial + '<br>' +
                    'State: ' + selectedDevice.state + '<br>' +
                    'Connection: ' + selectedDevice.connection;
                infoDiv.style.display = 'block';
                showStatus('Device selected', 'success');
            } catch (err) {
                showStatus('Failed to select device: ' + err.message, 'error');
            }
        });

        async function connectWiFi() {
            var ip = document.getElementById('wifiIP').value.trim();
            var port = parseInt(document.getElementById('wifiPort').value) || 5555;

            if (!ip) {
                showStatus('Please enter IP address', 'error');
                return;
            }

            showStatus('Connecting to ' + ip + ':' + port + '...', 'info');

            try {
                var res = await fetch('/api/connect', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ ip: ip, port: port })
                });
                var data = await res.json();

                if (data.success) {
                    showStatus('Connected successfully!', 'success');
                    refreshDevices();
                } else {
                    showStatus('Connection failed: ' + data.message, 'error');
                }
            } catch (err) {
                showStatus('Failed to connect: ' + err.message, 'error');
            }
        }

        async function disconnectWiFi() {
            var ip = document.getElementById('wifiIP').value.trim();
            var port = parseInt(document.getElementById('wifiPort').value) || 5555;

            if (!ip) {
                showStatus('Please enter IP address', 'error');
                return;
            }

            try {
                var res = await fetch('/api/disconnect', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ ip: ip, port: port })
                });
                var data = await res.json();

                if (data.success) {
                    showStatus('Disconnected successfully!', 'success');
                    refreshDevices();
                } else {
                    showStatus('Disconnect failed: ' + data.message, 'error');
                }
            } catch (err) {
                showStatus('Failed to disconnect: ' + err.message, 'error');
            }
        }

        async function sendBroadcast() {
            var content = document.getElementById('content').value.trim();
            var action = document.getElementById('action').value.trim();

            if (!content) {
                showStatus('Please enter broadcast content', 'error');
                return;
            }
            if (!selectedDevice) {
                showStatus('Please select a device', 'error');
                return;
            }

            try {
                var res = await fetch('/api/send', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ content: content, action: action })
                });
                var data = await res.json();

                if (data.success) {
                    showStatus('Broadcast sent successfully!', 'success');
                    document.getElementById('content').value = '';
                    loadHistory();
                } else {
                    showStatus('Broadcast failed: ' + (data.message || 'Unknown error'), 'error');
                }
            } catch (err) {
                showStatus('Failed to send broadcast: ' + err.message, 'error');
            }
        }

        async function loadHistory() {
            try {
                var res = await fetch('/api/history');
                var data = await res.json();
                var historyDiv = document.getElementById('history');

                if (data.data && data.data.length > 0) {
                    var html = '';
                    for (var i = 0; i < data.data.length; i++) {
                        var item = data.data[i];
                        var statusClass = item.status.toLowerCase();
                        var contentPreview = item.content.length > 100 ? item.content.substring(0, 100) + '...' : item.content;
                        html += '<div class="history-item ' + statusClass + '">' +
                            '<div class="timestamp">' + item.timestamp + '</div>' +
                            '<div class="content">' + escapeHtml(contentPreview) + '</div>' +
                            '<span class="status ' + statusClass + '">' + item.status + '</span>' +
                            (item.error ? '<div style="color:#C62828;font-size:12px;margin-top:5px;">' + escapeHtml(item.error) + '</div>' : '') +
                            '</div>';
                    }
                    historyDiv.innerHTML = html;
                } else {
                    historyDiv.innerHTML = '<div class="empty-state">No broadcast history</div>';
                }
            } catch (err) {
                console.error('Failed to load history:', err);
            }
        }

        // Quick content functions
        function setQuickContent(content) {
            var contentTextarea = document.getElementById('content');
            contentTextarea.value = content;
            contentTextarea.focus();
            showStatus('Content set: ' + (content.length > 50 ? content.substring(0, 50) + '...' : content), 'success');
        }

        function addCustomContent() {
            var customInput = document.getElementById('customContent');
            var content = customInput.value.trim();

            if (!content) {
                showStatus('Please enter content to add', 'error');
                return;
            }

            // Create new quick content button
            var buttonsContainer = document.querySelector('.quick-content-buttons');
            var newButton = document.createElement('button');
            newButton.className = 'btn-quick';
            newButton.textContent = content.length > 20 ? content.substring(0, 20) + '...' : content;
            newButton.title = content;
            newButton.onclick = function() {
                setQuickContent(content);
            };

            buttonsContainer.appendChild(newButton);

            // Clear input
            customInput.value = '';
            showStatus('Custom content added: ' + (content.length > 30 ? content.substring(0, 30) + '...' : content), 'success');
        }

        async function clearHistory() {
            try {
                await fetch('/api/clear-history', { method: 'POST' });
                showStatus('History cleared', 'success');
                loadHistory();
            } catch (err) {
                showStatus('Failed to clear history: ' + err.message, 'error');
            }
        }

        function showStatus(message, type) {
            var statusDiv = document.getElementById('status');
            statusDiv.textContent = message;
            statusDiv.className = 'status ' + type;
            statusDiv.style.display = 'block';
            setTimeout(function() { statusDiv.style.display = 'none'; }, 5000);
        }

        function escapeHtml(text) {
            var div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
    </script>
</body>
</html>`
