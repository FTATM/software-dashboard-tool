# WSL2 Server Deployment & Networking Guide

Complete setup guide for hosting an unattended Docker IoT stack in WSL2 on Windows 10/11: managing RAM consumption, keeping WSL running permanently in the background without UI windows, exposing required services across the LAN, and securing local databases.

---

## 1. Network & Port Architecture

| Service | Port | Protocol | LAN Exposure | Access Endpoint |
| :--- | :--- | :--- | :--- | :--- |
| **Frontend UI** | `80` | TCP | **Allowed** | `http://<WINDOWS_IP>` |
| **Node-RED** | `1880` | TCP | **Allowed** | `http://<WINDOWS_IP>:1880` |
| **Mosquitto MQTT** | `1883` | TCP | **Allowed** | `<WINDOWS_IP>:1883` |
| **Device Gateway** | `8090` | TCP | **Allowed** | `<WINDOWS_IP>:8090` |
| **Device Gateway API** | `8091` | TCP | **Allowed** | `http://<WINDOWS_IP>:8091` |
| **TimescaleDB** | `5433` | TCP | **Blocked (Host Only)** | `localhost:5433` on Windows host only |

---

## 2. RAM & Resource Optimization

WSL2 allocates up to 50% of total host RAM by default, and Linux caches disk files in RAM without releasing them, causing Windows `Vmmem` to spike significantly.

### Step 2.1: Enforce Memory Ceilings via `.wslconfig`
Create or edit `.wslconfig` in your Windows user profile folder:
1. Press **`Win + R`**, type `notepad %USERPROFILE%\.wslconfig`, and press **Enter**.
2. Add the following:

```ini
[wsl2]
# Restricts maximum RAM allocated to the WSL VM
memory=1GB

# Emergency disk swap to absorb workload spikes and avoid OOM crashes
swap=2GB

# Automatically releases reclaimed Linux page cache memory back to Windows
autoMemoryReclaim=gradual
```

3. Save the file and restart WSL to apply:
   ```powershell
   wsl --shutdown
   ```

### (optional) Constrain Service Memory in `docker-compose.yml`
Apply container limits to prevent database buffers and the Node-RED runtime from growing unchecked:

```yaml
services:
  # Limit TimescaleDB/PostgreSQL shared memory buffers
  db:
    image: timescale/timescaledb:2.29.2-pg18
    mem_limit: 384m
    command: postgres -c shared_buffers=64MB -c work_mem=4MB -c max_connections=40

  # Cap Node-RED V8 engine heap
  nodered:
    image: nodered/node-red:5.0-debian
    mem_limit: 256m
    environment:
      - TZ=Asia/Bangkok
      - NODE_OPTIONS=--max_old_space_size=160

  # Microservices and proxies
  server-api:
    mem_limit: 256m
  server-schedule-engine:
    mem_limit: 256m
  server-device-gateway:
    mem_limit: 256m
  frontend:
    mem_limit: 64m
  mqtt-broker:
    mem_limit: 64m
  garage:
    mem_limit: 192m
```

### Step 2.2: Reclaiming Cache on Demand
To immediately flush buffered Linux page cache without restarting WSL:
```bash
sudo sync && sudo sh -c "echo 3 > /proc/sys/vm/drop_caches"
```

---

## 3. Persistent WSL Background Execution

WSL automatically terminates the entire virtual machine 15 seconds after all interactive terminal windows are closed. Running an invisible VBScript at startup maintains an open background process (`sleep infinity`) that keeps the VM alive 24/7 with zero desktop windows.

1. Press **`Win + R`**, type `shell:startup`, and press **Enter**.
2. Create a file named `wsl_background.vbs`.
3. Open the file in Notepad and add:
   ```vbscript
   CreateObject("Wscript.Shell").Run "wsl.exe -d Ubuntu --exec sleep infinity", 0, False
   ```
4. Save and close. Double-click `wsl_background.vbs` once to initiate persistent background execution immediately.

---

## 4. Windows Firewall Configuration

Run the following command once in **PowerShell (Run as Administrator)** on the Windows host to allow incoming LAN traffic to the exposed service ports:

```powershell
New-NetFirewallRule -DisplayName "WSL IoT Stack LAN Access" `
  -Direction Inbound `
  -LocalPort 80,1880,1883,8090,8091 `
  -Protocol TCP `
  -Action Allow
```

*(Port `5433` is intentionally omitted so TimescaleDB remains unexposed to the local network).*

---

## 5. Automated Port Forwarding Script

Because WSL2 operates inside an isolated NAT virtual network, its internal IP address changes across reboots. This script queries the active WSL IP address and updates Windows `netsh portproxy` dynamically.

1. Create folder: `C:\scripts\`
2. Save script as: `C:\scripts\wsl_portproxy.ps1`

```powershell
# wsl_portproxy.ps1
# Wait until WSL boots and exposes a valid internal IP
$wslIp = ""
while (-not $wslIp) {
    Start-Sleep -Seconds 2
    $wslIp = (wsl -d Ubuntu hostname -I).Trim().Split(" ")[0]
}

# Define LAN-accessible TCP ports (TimescaleDB 5433 omitted for host security)
$ports = @(80, 1880, 1883, 8090, 8091)

# Reset stale port mappings
netsh interface portproxy reset

# Register new proxy rules mapped to the current WSL IP
foreach ($port in$ports) {
    netsh interface portproxy add v4tov4 `
        listenport=$port `
        listenaddress=0.0.0.0 `
        connectport=$port `
        connectaddress=$wslIp
}

# Restart Windows IP Helper service to apply changes cleanly
Restart-Service iphlpsvc
```

---

## 6. Task Scheduler Automation

Because modifying network routing tables with `netsh` requires administrator privileges and Windows accounts without passwords cannot utilize *"Run whether user is logged on or not"*, use a Logon-triggered elevated task.

1. Press **`Win + R`**, type `taskschd.msc`, and press **Enter**.
2. Select **Task Scheduler Library** > click **Create Task...** (in the right panel).
3. Configure each tab:

* **General Tab:**
  * **Name:** `WSL Port Forwarding`
  * Select: **Run only when user is logged on**
  * Check: **Run with highest privileges** *(Required for network routing)*
  * **Configure for:** *Windows 10*
* **Triggers Tab:**
  * Click **New...**
  * **Begin the task:** *At log on*
  * **Settings:** Any user or your current local user account
  * Click **OK**
* **Actions Tab:**
  * Click **New...**
  * **Action:** *Start a program*
  * **Program/script:** `powershell.exe`
  * **Add arguments:**
    ```text
    -WindowStyle Hidden -ExecutionPolicy Bypass -File "C:\scripts\wsl_portproxy.ps1"
    ```
  * Click **OK**
4. Click **OK** to save the task.
5. In **Task Scheduler Library**, right-click **`WSL Port Forwarding`** and select **Run** to execute the script immediately.

---

## 7. UDP Protocol Note & Device Gateway

* **TCP vs. UDP Routing:** Windows `netsh interface portproxy` routes **TCP only**. It cannot forward UDP packets.
* **Netcat False Positives:** Testing with `nc -zvu <WINDOWS_IP> 8090` from a remote terminal can return `Connection succeeded!` even when ports are closed. This happens because Windows Firewall silently discards unauthorized UDP packets without returning ICMP "Port Unreachable" packets.
* **Raw UDP Requirements:** If physical devices require true UDP datagram delivery to port 8090, install a dedicated UDP forwarder (such as `socat` for Windows) and add a matching Windows Firewall UDP inbound rule.

---

## 8. Verification & Troubleshooting Commands

### On Windows Server (PowerShell)
```powershell
# Verify active port forwarding bindings
netsh interface portproxy show all

# Verify the WSL instance is running
wsl --list --running

# Test local web server response
curl http://localhost
```

### On Client Machine
```bash
# Verify Frontend HTTP response example as 192.168.1.138 is server
curl -I http://192.168.1.138

# Test connectivity to exposed TCP ports
nc -zv 192.168.1.138 80      # Frontend UI
nc -zv 192.168.1.138 1880    # Node-RED
nc -zv 192.168.1.138 1883    # Mosquitto MQTT
nc -zv 192.168.1.138 8090    # Device Gateway (TCP)
nc -zv 192.168.1.138 8091    # Device Gateway API

# Verify TimescaleDB is blocked from LAN access
nc -zv -w 2 192.168.1.138 5433
# (Expected output: Connection refused or timed out)
```