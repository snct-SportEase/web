#!/usr/bin/env python3
"""Aggregate Gin request logs from stdin without including user IDs or cookies."""
import json
import math
import re
import sys

routes = {"/api/auth/user", "/api/events/active", "/api/student/class-progress", "/api/scores/class"}
pattern = re.compile(r'\|\s*(\d+)\s*\|\s*([\d.]+)(ns|µs|us|ms|s)\s*\|.*\|\s*GET\s+"([^"?]+)')
scale = {"ns": 1e-6, "µs": 1e-3, "us": 1e-3, "ms": 1, "s": 1000}
samples = {}
for line in sys.stdin:
    match = pattern.search(line)
    if not match:
        continue
    status, duration, unit, path = match.groups()
    if path not in routes:
        continue
    samples.setdefault((path, status), []).append(float(duration) * scale[unit])
for (path, status), values in sorted(samples.items()):
    values.sort()
    print(json.dumps({"path": path, "status": int(status), "count": len(values),
                      "mean_ms": round(sum(values) / len(values), 3),
                      "p95_ms": round(values[math.ceil(len(values) * .95) - 1], 3)}))
