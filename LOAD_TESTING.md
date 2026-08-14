# Load testing (k6)

Target: **20 000 RPS for 1 minute** against the posts list page (`GET /forum/?page=N`, N ∈ {1,2,3}) of the Docker container on `localhost:8084` (`POSTS_CACHE_TTL=5s`). Host: 16-core Linux, k6 v2.2.0.

```bash
DURATION=60s RPS_LEVELS=20000 bash k6/run-load.sh posts
```

## Results

| Metric | Value |
|---|---|
| Target RPS | 20 000 |
| Duration | 60 s |
| Requests sent | 1 194 522 (19 908.2 req/s sustained) |
| Dropped iterations | 5 479 (91.3/s, ~0.46% of target) |
| Success (HTTP 200) | 100.00% |
| Latency avg / med | 2.32 ms / 0.44 ms |
| Latency p90 / p95 / max | 5.22 ms / 12.01 ms / 124.74 ms |
| Data received | 14 GB (228 MB/s) |
| Max VUs | 692 |

### CPU usage (sampled every 5 s, % of one core)

The forum app used ~4.5 of the 16 cores, k6 ~5; drops were negligible (~0.46%).

| Time | Forum app | k6 |
|---|---|---|
| 02:15:55 | 429.3% | 494.8% |
| 02:16:00 | 445.7% | 496.1% |
| 02:16:05 | 446.4% | 500.5% |
| 02:16:10 | 447.8% | 495.5% |
| 02:16:15 | 447.4% | 504.1% |
| 02:16:20 | 443.0% | 488.7% |
| 02:16:25 | 451.2% | 502.9% |
| 02:16:30 | 451.6% | 501.0% |
| 02:16:36 | 450.6% | 500.1% |
| 02:16:41 | 441.6% | 495.1% |
| 02:16:46 | 448.1% | 504.2% |
| **avg** | **445.7%** | **498.5%** |