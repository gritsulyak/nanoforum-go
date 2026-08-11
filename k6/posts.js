import http from 'k6/http';
import { check } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8084';
const BASE_PATH = __ENV.BASE_PATH || '/forum';
const TARGET_RPS = Number(__ENV.RPS || 5000);
const DURATION = __ENV.DURATION || '10s';
const PAGES = (__ENV.PAGES || '1,2,3').split(',').map(Number).filter((n) => n > 0);
const PRE_ALLOCATED_VUS = Number(__ENV.PRE_ALLOCATED_VUS || 500);
const MAX_VUS = Number(__ENV.MAX_VUS || 20000);
const LIST_URL = `${BASE_URL}${BASE_PATH}/`;

const failures = new Rate('posts_failures');
const latency = new Trend('posts_duration');

export const options = {
  scenarios: {
    posts: {
      executor: 'constant-arrival-rate',
      rate: TARGET_RPS,
      timeUnit: '1s',
      duration: DURATION,
      preAllocatedVUs: PRE_ALLOCATED_VUS,
      maxVUs: MAX_VUS,
    },
  },
  thresholds: {
    posts_failures: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
};

export default function () {
  const page = PAGES[Math.floor(Math.random() * PAGES.length)];
  const res = http.get(`${LIST_URL}?page=${page}`);

  const ok = res.status === 200;
  failures.add(!ok);
  latency.add(res.timings.duration);

  check(res, {
    'posts page served (200)': () => ok,
  });
}
