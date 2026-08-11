import http from 'k6/http';
import { check } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8084';
const BASE_PATH = __ENV.BASE_PATH || '/forum';
const TARGET_RPS = Number(__ENV.RPS || 5000);
const DURATION = __ENV.DURATION || '10s';
const PRE_ALLOCATED_VUS = Number(__ENV.PRE_ALLOCATED_VUS || 500);
const MAX_VUS = Number(__ENV.MAX_VUS || 20000);
const USERNAME = __ENV.USERNAME || 'loadtest';
const PASSWORD = __ENV.PASSWORD || 'loadtest';
const LOGIN_URL = `${BASE_URL}${BASE_PATH}/login`;

const failures = new Rate('login_failures');
const latency = new Trend('login_duration');

export const options = {
  scenarios: {
    login: {
      executor: 'constant-arrival-rate',
      rate: TARGET_RPS,
      timeUnit: '1s',
      duration: DURATION,
      preAllocatedVUs: PRE_ALLOCATED_VUS,
      maxVUs: MAX_VUS,
    },
  },
  thresholds: {
    login_failures: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
};

export default function () {
  const payload = `username=${encodeURIComponent(USERNAME)}&password=${encodeURIComponent(PASSWORD)}`;
  const res = http.post(LOGIN_URL, payload, {
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    redirects: 0,
  });

  const ok = res.status === 303 || res.status === 302;
  failures.add(!ok);
  latency.add(res.timings.duration);

  check(res, {
    'login succeeded (303/302 redirect)': () => ok,
  });
}
