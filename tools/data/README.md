# 📂 Hoverfly Smoke-Test Data

This directory contains the raw API curl command dump used by the test runner.

## 📝 curl_dump.txt

This file contains all the JSON-RPC curl request examples extracted from the official Hive Developer documentation. By default, the test runner reads this file and fires each command against the local mock server to verify the payload structures, parameters, and response keys.

---

## 🛠️ Generation Process

If you need to regenerate or update `curl_dump.txt` from the live documentation, you can use the following commands:

```bash
# 1. Fetch the raw apidefinitions page, strip html tags (using strip-tags or similar utility),
#    grep for curl commands, and save to a temporary dump file
curl -fsSL https://developers.hive.io/apidefinitions/ | strip-tags code | grep curl > curl_dump.txt

# 2. Insert newlines before each curl command to clean up the dump format
sed -i 's/./\nc/' curl_dump.txt

# 3. Swap out the production Hive node URL (https://api.hive.blog) for the local Hoverfly mock server URL
sed -i 's|https://api.hive.blog|http://localhost:8090|' curl_dump.txt
```

> [!NOTE]
> The [strip-tags](https://pypi.org/project/strip-tags/) utility is a CLI tool used to remove HTML tags. You can use any similar parser or python equivalent if `strip-tags` is not available on your system.
