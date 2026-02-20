
# SNMP Mock & Go Client

This project provides an isolated, local testing environment for SNMP. It uses a Python-based SNMP simulator (`snmpsim`) to mock network devices (like Cisco or Juniper routers) and a custom Go client to query that mock data using `BulkWalk`.

By containerizing both the mock server and the client with Docker Compose, you can safely test SNMP polling logic without needing access to real, physical network hardware.

## 📂 Repository Structure

```text
.
├── docker-compose.yml        
├── .env.example              
├── .gitignore                
├── data/
│   └── test.snmpwalk.example 
└── go-client/
    ├── Dockerfile            
    ├── go.mod               
    ├── go.sum                
    └── main.go               
```

## ⚙️ Prerequisites

* [Docker](https://docs.docker.com/get-docker/)
* [Docker Compose](https://docs.docker.com/compose/install/)
* `snmpwalk` installed on your host machine (optional, but highly recommended for manual testing).

---

## 🚀 Quick Start Guide

### 1. Set Up Your Environment Variables

We use a `.env` file to pass configuration into the Docker containers. Copy the provided `.env.example` template to create your active, git-ignored `.env` file:

```bash
cp .env.example .env
```

Update the `.env` file with your desired target OID and community string.

### 2. Set Up Your Mock Data

The Python simulator loads mock data from `.snmpwalk` or `.snmprec` files inside the `data/` directory.

**🚨 CRUCIAL NOTE:** The name of the file dictates the **SNMP Community String**.

Copy the example data file to create your active mock data. For example, if you set `SNMP_COMMUNITY=cisco_switch` in your `.env` file, you must name the file `cisco_switch.snmpwalk`:

```bash
cp data/test.snmpwalk.example data/cisco_switch.snmpwalk
```

### 3. Run the Stack

Spin up the entire environment using Docker Compose. The `--build` flag ensures your Go client is freshly compiled with any recent code changes.

```bash
docker compose up --build
```

**What happens next:**

1. **Boot:** The `snmp-mock` container starts and loads your `.snmpwalk` file into memory.
2. **Wait:** The `snmp-client` (Go) container boots and uses a retry loop to ping the mock server, waiting for UDP port 161 to open.
3. **Walk:** Once connected, the Go client executes a `BulkWalk` against your configured OID.
4. **Exit:** The client parses and prints the OID data to your terminal, then gracefully exits.

---

## 🧪 Manual Testing

If you want to test the mock server directly from your local terminal (bypassing the Go client entirely), keep the mock server running in the background:

```bash
docker compose up -d snmp-mock
```

Then, run a standard `snmpwalk` command against your local machine (`127.0.0.1`). Replace `cisco_switch` with your filename, and the trailing OID with the subtree you want to test:

```bash
snmpwalk -v2c -c cisco_switch 127.0.0.1 .1.3.6.1.2.1.15.6.1
```

---

## 🛠️ Environment Variables Reference

| Variable           | Default / Example       | Description                                                                       |
| :----------------- | :---------------------- | :-------------------------------------------------------------------------------- |
| `SNMP_TARGET`    | `snmp-mock`           | The hostname of the SNMP server (matches the Compose service name).               |
| `SNMP_PORT`      | `161`                 | The UDP port the SNMP server is listening on.                                     |
| `SNMP_COMMUNITY` | `cisco_switch`        | The SNMP community string.**Must exactly match the filename** in `/data`. |
| `SNMP_OID`       | `.1.3.6.1.2.1.15.6.1` | The parent OID subtree the Go client will BulkWalk.                               |

## 🐛 Troubleshooting

* **`Connection Refused` or Go script crashes:** Ensure the Go client is using its retry loop. Because SNMP uses UDP, the Go script might fire packets before the Python server has fully parsed the text file and opened the port.
* **`Timeout: No Response` during manual test:** Your community string does not match the filename in the `data/` folder, or the mock server failed to read the file due to formatting errors. Check the logs: `docker logs snmp-mock`.
