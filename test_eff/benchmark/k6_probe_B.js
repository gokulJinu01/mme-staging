import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 5,
  duration: '30s',
  thresholds: {
    http_req_duration: ['p(95)<120'],
    http_req_failed: ['rate<0.01'],
  },
};

export default function () {
  const url = 'http://localhost:8081/search/semantic';
  const params = {
    headers: {
      'X-User-ID': 'benchmark_user',
      'Content-Type': 'application/json',
    },
  };
  
  const payload = {
    query: 'machine learning algorithms'
  };

  const response = http.post(url, JSON.stringify(payload), params);
  
  check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 120ms': (r) => r.timings.duration < 120,
    'has success field': (r) => r.json('success') === true,
    'has data field': (r) => r.json('data') !== undefined,
    'has meta field': (r) => r.json('meta') !== undefined,
  });

  sleep(1);
}
