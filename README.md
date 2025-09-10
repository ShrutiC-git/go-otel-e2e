# 🚀 End-to-End Observability in Go with OTel and SigNoz

- [🚀 End-to-End Observability in Go with OTel and SigNoz](#-end-to-end-observability-in-go-with-otel-and-signoz)
  - [Prerequisites](#prerequisites)
  - [Setup Instructions](#setup-instructions)
    - [1. Clone the repository](#1-clone-the-repository)
    - [2. Install Go Dependencies](#2-install-go-dependencies)
    - [3. Prepare the OpenTelemetry Collector Config File](#3-prepare-the-opentelemetry-collector-config-file)
    - [4. Run the OpenTelemetry Collector](#4-run-the-opentelemetry-collector)
    - [5. Run the Go Service](#5-run-the-go-service)
      - [Test the order endpoint:](#test-the-order-endpoint)
      - [Test the inventory endpoint:](#test-the-inventory-endpoint)
  - [Architecture Overview](#architecture-overview)
  - [Troubleshooting](#troubleshooting)
    - [1. Traces, Metrics, or Logs are Not Appearing in SigNoz](#1-traces-metrics-or-logs-are-not-appearing-in-signoz)
      - [Is the OpenTelemetry Collector Running?](#is-the-opentelemetry-collector-running)
      - [Is the Collector Configured Correctly?](#is-the-collector-configured-correctly)
      - [Is the Go App Connecting to the Collector?](#is-the-go-app-connecting-to-the-collector)
    - [2. The Go Service Fails to Start](#2-the-go-service-fails-to-start)
      - [Did you run go mod tidy?](#did-you-run-go-mod-tidy)
  - [Next Steps](#next-steps)


This project demonstrates how to capture **traces, metrics, and logs** from HTTP-based microservices in Go, send them to an **OpenTelemetry Collector running locally**, and visualize them in **SigNoz Cloud**.

![Go](https://img.shields.io/badge/Go-1.23+-blue?logo=go) 
![OpenTelemetry](https://img.shields.io/badge/OpenTelemetry-v1.0-purple) 
![SigNoz](https://img.shields.io/badge/Backend-SigNoz-orange)

---

## Prerequisites

- [Go 1.23+](https://go.dev/dl/)
- [SigNoz Cloud account](https://signoz.io) (or self-hosted SigNoz)
- [SigNoz Ingestion Key](https://signoz.io/docs/ingestion/signoz-cloud/keys/)
- [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector-releases) 
  
> **Note:**  
> This project was built and tested on **Windows**, with the OpenTelemetry Collector running locally on Windows.  
> You can also run the Collector on **Linux** or **macOS**, or use a **Docker container** instead of the native binary. The configuration and telemetry flow remain the same across platforms.


---

## Setup Instructions

### 1. Clone the repository

```bash
git clone https://github.com/ShrutiC-git/go-otel-e2e.git
cd go-otel-e2e
```

### 2. Install Go Dependencies

```bash
go mod tidy
```

This will download all dependencies required to run the application.

### 3. Prepare the OpenTelemetry Collector Config File

1. Copy the provided [`otel-collector-config.yaml`](/otel-collector-config.yaml) into the same folder as the OpenTelemetry Collector binary.  
2. Open `config.yaml` in a text editor.  
3. Update the following fields to point to your SigNoz Cloud account:

| Field                  | Description                                | Example / Notes                                     |
|------------------------|--------------------------------------------|----------------------------------------------------|
| `endpoint`             | OTLP endpoint for sending telemetry data  | `ingest.<YOUR_SIGNOZ_REGION>.signoz.cloud:443`   |
| `signoz-ingestion-key` | Your SigNoz ingestion key                 | `<YOUR_SIGNOZ_INGESTION_KEY>`                     |

4. Save the file. 

> Once configured, the Collector will export traces, metrics, and logs from your Go services to SigNoz Cloud.

### 4. Run the OpenTelemetry Collector

1. Open **PowerShell** or **Command Prompt**.  
2. Navigate to the folder containing `otelcol-contrib.exe`.  
3. Run the Collector with the config file:

```powershell
.\otelcol-contrib.exe --config config.yaml
```

If your config file is in a different path:

```powershell
.\otelcol-contrib.exe --config <path-to-config.yaml>
```

You should see `Health Check is ready` in the logs, indicating the Collector is running.

> **Note**: otelcol-contrib is the name of the OpenTelemetry Collector executable. This specific version includes additional exporters and receivers needed for this project.

### 5. Run the Go Service

Start the Go application:

```bash
go run main.go
```

The service will start on port `8080` and expose two sample endpoints:  

- `http://localhost:8080/createOrder`  
- `http://localhost:8080/checkInventory`  

#### Test the order endpoint:
```bash
curl http://localhost:8080/createOrder
```

#### Test the inventory endpoint:
```bash
curl http://localhost:8080/checkInventory
```

> Once the OpenTelemetry Collector is running, all requests to these endpoints will be recorded, and the traces, metrics, and logs will be visible in **SigNoz Cloud**.

---

## Architecture Overview

```
+------------------+       +----------------------+       +----------------+
|  Go Microservice | ----> | OTel Collector (local)| ---> |   SigNoz Cloud |
| (with OTel SDK)  |       |  Receives traces via |       | (Backend + UI) |
|                  |       |  OTLP (http/grpc)    |       |                |
+------------------+       +----------------------+       +----------------+
```

---

## Troubleshooting

If you're having issues, here are a few common problems and their solutions.

### 1. Traces, Metrics, or Logs are Not Appearing in SigNoz
This is usually caused by a connectivity or configuration issue.

#### Is the OpenTelemetry Collector Running?

Confirm that the Collector binary is running in your terminal and you see the Health Check is ready message. If not, double-check that you've navigated to the correct directory and run the command from Step 4 correctly. 

Alternatively, you can also check `Windows Service` to check the `OTel Collector` is running.

#### Is the Collector Configured Correctly?

Verify that your otel-collector-config.yaml file contains the correct endpoint and signoz-ingestion-key for your SigNoz Cloud account. A simple typo in the ingestion key can prevent data from being ingested. Check `Windows Application Logs` if issues persist.

#### Is the Go App Connecting to the Collector?

The Go service is configured to send telemetry to localhost:4318. Make sure your local OpenTelemetry Collector is running and listening on this address and port.

Check for any firewall rules that might be blocking the connection between your Go application and the Collector.

### 2. The Go Service Fails to Start
If the application won't start, it's often a dependency issue.

#### Did you run go mod tidy?

Ensure you have run go mod tidy after cloning the repository to download all necessary dependencies. Check your terminal output for any errors during this process.

---

## Next Steps

- Explore the traces in the **SigNoz UI** under the *Traces* tab.  
- Check out the *Metrics* and *Logs* dashboards.  