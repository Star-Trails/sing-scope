================================================================================
  sing-scope: Cross-Platform sing-box Traffic Analyzer (Portable Edition)
================================================================================

QUICK START INSTRUCTIONS:
-------------------------
1. Ensure your sing-box instance (v1.14.0+) is running with the new API service:
   Example sing-box config snippet:
   {
     "services": [
       {
         "type": "api",
         "tag": "api",
         "listen": "127.0.0.1",
         "listen_port": 9090,
         "secret": "my-secret-token"
       }
     ]
   }

2. Double-click TrafficAnalyzer.exe to launch the application.
   - The embedded UI will automatically open in your default web browser.
   - Default address: http://127.0.0.1:19876

3. In the Settings view or top bar:
   - Enter your sing-box API Server URL (e.g. http://127.0.0.1:9090)
   - Enter the Secret (if configured)
   - Click "Connect"

4. Generate network traffic through your sing-box TUN inbound to observe
   live connections, transfer rates, process attribution, and routing breakdowns.

COMMAND-LINE OPTIONS:
---------------------
  -singbox-url string
        Default sing-box API server URL (default: "http://127.0.0.1:9090")
  -secret string
        Default sing-box API secret / Bearer token
  -port int
        Local port to bind the embedded UI server to (default: 19876)
  -host string
        Local host to bind UI server to (default: "127.0.0.1")
  -no-browser
        Do not automatically open the web browser on launch

SYSTEM REQUIREMENTS:
--------------------
- Windows 10 / 11 (x86_64)
- No installer or external runtime dependencies required.
- Configuration is stored locally in memory or user configuration directory.
================================================================================
