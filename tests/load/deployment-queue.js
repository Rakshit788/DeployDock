import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter } from 'k6/metrics';

const baseURL = __ENV.BASE_URL || 'http://api:8080';
const projectID = __ENV.TEST_PROJECT_ID;

if (!projectID) throw new Error('TEST_PROJECT_ID is required. This test creates real deployment jobs.');

export const options = {
  scenarios: {
    deployment_queue: {
      executor: 'per-vu-iterations', vus: Number(__ENV.VUS || 4),
      iterations: Number(__ENV.ITERATIONS_PER_VU || 1), maxDuration: '2m',
    },
  },
  thresholds: { http_req_failed: ['rate<0.01'], http_req_duration: ['p(95)<1000'], checks: ['rate>0.99'] },
};

const acceptedDeployments = new Counter('accepted_deployments');

export function setup() {
  // Get test token before running tests
  const tokenResponse = http.get(`${baseURL}/api/test/token`);
  if (tokenResponse.status !== 200) {
    throw new Error(`Failed to get test token: ${tokenResponse.status} ${tokenResponse.body}`);
  }
  const tokenData = tokenResponse.json();
  return { token: tokenData.token };
}

export default function (data) {
  const response = http.post(`${baseURL}/create-deployment`, JSON.stringify({ project_id: Number(projectID) }), {
    headers: { 
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${data.token}`
    }, 
    tags: { endpoint: 'create_deployment' },
  });
  const accepted = check(response, {
    'deployment is accepted': (res) => res.status === 200,
    'response includes deployment ID': (res) => {
      try { return Boolean(res.json('DEPLOYMENT_ID')); } catch (_) { return false; }
    },
  });
  if (accepted) acceptedDeployments.add(1);
  sleep(1);
}
