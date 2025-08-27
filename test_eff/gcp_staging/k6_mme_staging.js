import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 20 },  // Ramp up to 20 VUs
    { duration: '2m', target: 20 },   // Stay at 20 VUs for 2 minutes
    { duration: '30s', target: 0 },   // Ramp down to 0 VUs
  ],
  thresholds: {
    http_req_duration: ['p(95)<120'],  // 95% of requests must complete below 120ms
    http_req_failed: ['rate<0.01'],    // Error rate must be less than 1%
  },
};

const BASE_URL = 'http://34.58.167.157';
const MME_HOST = 'mme.34.58.167.157.nip.io';
const TAGMAKER_HOST = 'tagmaker.34.58.167.157.nip.io';

export default function () {
  const params = {
    headers: {
      'Host': MME_HOST,
      'X-User-ID': 'load_test_user',
      'Content-Type': 'application/json',
    },
  };

  // Test 1: MME Health Check
  const healthResponse = http.get(`${BASE_URL}/health`, params);
  check(healthResponse, {
    'MME health check status is 200': (r) => r.status === 200,
    'MME health check response time < 50ms': (r) => r.timings.duration < 50,
  });

  // Test 2: MME Memory Query
  const queryResponse = http.get(`${BASE_URL}/memory/query?tags=demo&limit=5`, params);
  check(queryResponse, {
    'MME query status is 200': (r) => r.status === 200,
    'MME query response time < 120ms': (r) => r.timings.duration < 120,
    'MME query has success field': (r) => r.json('success') === true,
    'MME query has data field': (r) => r.json('data') !== undefined,
    'MME query has meta field': (r) => r.json('meta') !== undefined,
  });

  // Test 3: MME Semantic Search
  const semanticParams = {
    headers: {
      'Host': MME_HOST,
      'X-User-ID': 'load_test_user',
      'Content-Type': 'application/json',
    },
  };
  
  const semanticPayload = JSON.stringify({
    query: 'machine learning algorithms',
    limit: 5
  });
  
  const semanticResponse = http.post(`${BASE_URL}/search/semantic`, semanticPayload, semanticParams);
  check(semanticResponse, {
    'MME semantic search status is 200': (r) => r.status === 200,
    'MME semantic search response time < 150ms': (r) => r.timings.duration < 150,
    'MME semantic search has success field': (r) => r.json('success') === true,
    'MME semantic search has data field': (r) => r.json('data') !== undefined,
  });

  // Test 4: Tagmaker Health Check
  const tagmakerParams = {
    headers: {
      'Host': TAGMAKER_HOST,
      'X-User-ID': 'load_test_user',
      'Content-Type': 'application/json',
    },
  };
  
  const tagmakerHealthResponse = http.get(`${BASE_URL}/health`, tagmakerParams);
  check(tagmakerHealthResponse, {
    'Tagmaker health check status is 200': (r) => r.status === 200,
    'Tagmaker health check response time < 50ms': (r) => r.timings.duration < 50,
  });

  // Test 5: MME Memory Save (with idempotency)
  const saveParams = {
    headers: {
      'Host': MME_HOST,
      'X-User-ID': 'load_test_user',
      'Content-Type': 'application/json',
      'Idempotency-Key': `test-${Date.now()}-${Math.random()}`,
    },
  };
  
  const savePayload = JSON.stringify({
    content: 'Load test memory entry',
    tags: [
      { label: 'load-test', type: 'test' },
      { label: 'performance', type: 'category' }
    ],
    status: 'completed'
  });
  
  const saveResponse = http.post(`${BASE_URL}/memory/save`, savePayload, saveParams);
  check(saveResponse, {
    'MME save status is 200': (r) => r.status === 200,
    'MME save response time < 200ms': (r) => r.timings.duration < 200,
    'MME save has success field': (r) => r.json('success') === true,
  });

  // Random sleep between requests to simulate real user behavior
  sleep(Math.random() * 2 + 1); // Sleep between 1-3 seconds
}
