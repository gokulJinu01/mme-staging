import http from "k6/http";
import { check, sleep } from "k6";
export const options = { vus: 20, duration: "30s", thresholds: { http_req_duration: ["p(95)<120"], http_req_failed: ["rate<0.01"] } };
export default function () {
  const url = __ENV.BASE + "/memory/query?tags=demo&limit=10";
  const headers = { "X-User-ID": __ENV.USER };
  const res = http.get(url, { headers });
  check(res, { "200": r => r.status === 200, "envelope": r => JSON.parse(r.body).success === true });
  sleep(0.05);
}
