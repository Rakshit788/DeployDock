import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend } from 'k6/metrics';

const baseURL = __ENV.BASE_URL || 'http://api:8080';

export const options = {
  scenarios: {
    baseline: {
      executor: 'ramping-vus', startVUs: 0,
      stages: [{ duration: '15s', target: 5 }, { duration: '30s', target: 5 }, { duration: '15s', target: 0 }],
      gracefulRampDown: '10s',
    },
  },
  thresholds: { http_req_failed: ['rate<0.01'], http_req_duration: ['p(95)<500'], checks: ['rate>0.99'] },
};

const failedChecks = new Counter('failed_checks');
const healthLatency = new Trend('health_latency', true);

function expectOK(response, name) {
  const passed = check(response, {
    [`${name} returns 200`]: (res) => res.status === 200,
    [`${name} returns JSON`]: (res) => (res.headers['Content-Type'] || '').includes('application/json'),
  });
  if (!passed) failedChecks.add(1);
}

export function setup() {
  const tokenRes = http.get(`${baseURL}/api/test/token`);
  if (tokenRes.status !== 200) {
    throw new Error(`Failed to get test token: ${tokenRes.status}`);
  }
  const token = tokenRes.json('token');
  return { headers: { Authorization: `Bearer ${token}` } };
}

export default function (data) {
  const health = http.get(`${baseURL}/health`, { tags: { endpoint: 'health' } });
  healthLatency.add(health.timings.duration);
  expectOK(health, 'health');
  expectOK(http.get(`${baseURL}/api/project/list`, { headers: data.headers, tags: { endpoint: 'project_list' } }), 'project list');
  expectOK(http.get(`${baseURL}/api/deployment/list`, { headers: data.headers, tags: { endpoint: 'deployment_list' } }), 'deployment list');
  sleep(1);
}
