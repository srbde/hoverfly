# 🛸 Hoverfly Curl Test & Stress Tool

This directory contains the automated JSON-RPC smoke-test and verification tool for the **Hoverfly** mock Hive blockchain server. It extracts and runs the official API curl command examples from the [Hive Developer Portal](https://developers.hive.io) to ensure 100% API compliance and robustness.

## 🚀 Getting Started

The tool is written in Python and is optimized to run with **`uv`**, the ultra-fast Python package and project manager.

### Running the Tests

To run the complete test suite against a running Hoverfly instance (defaulting to `http://localhost:8090`):

```bash
# Make sure Hoverfly is running first
# cd hoverfly && go run .

# Run the test suite sequentially
uv run test_hoverfly.py
```

### Parallel Execution (Stress Test Mode)

To speed up execution or perform load/stress testing on the mock server, you can use the `--parallel` option to run requests concurrently:

```bash
# Run with 4 parallel worker threads
uv run test_hoverfly.py --parallel 4
```

---

## 🔧 Command-Line Options

```text
Usage:
    uv run test_hoverfly.py [options]

Options:
    --url TEXT       Base URL of the Hoverfly instance (default: http://localhost:8090)
    --dump FILE      Path to the curl dump file        (default: data/curl_dump.txt)
    --fail-only      Only print failing tests to the console
    --json           Write execution results to hoverfly_test_results.json
    --timeout FLOAT  Per-request timeout in seconds    (default: 10.0)
    --parallel INT   Number of parallel workers        (default: 1, sequential)
```

---

## 📊 Test Data

The underlying test data is located in the [data/](/data) folder. See [data/README.md](data/README.md) for details on how the raw curl commands were extracted and processed.
