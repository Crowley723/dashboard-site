import http from "k6/http";
import { check, sleep } from "k6";
import { Rate } from "k6/metrics";

const errorRate = new Rate("errors");

export const options = {
  stages: [
    { duration: "30s", target: 10 }, // Ramp up to 10 users
    { duration: "1m", target: 10 }, // Stay at 10 users
    { duration: "30s", target: 25 }, // Ramp up to 25 users
    { duration: "2m", target: 25 }, // Stay at 25 users
    { duration: "30s", target: 0 }, // Ramp down to 0
  ],
  thresholds: {
    http_req_duration: ["p(95)<500"], // 95% of requests must complete below 500ms
    errors: ["rate<0.1"], // Error rate must be below 10%
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";

function heavyDashboardUser() {
  let res = http.get(`${BASE_URL}/api/auth/status`);
  check(res, {
    "auth status check": (r) => r.status === 200 || r.status === 401,
  }) || errorRate.add(1);

  sleep(0.5);

  res = http.get(`${BASE_URL}/api/queries`);
  let queries = [];
  const queriesSuccess = check(res, {
    "queries retrieved": (r) => r.status === 200,
  });

  if (queriesSuccess && res.json()) {
    queries = res.json();
  } else {
    errorRate.add(1);
  }

  sleep(1);

  if (queries.length > 0) {
    const selectedQueries = queries.slice(0, Math.min(5, queries.length));
    res = http.get(`${BASE_URL}/api/data?queries=${selectedQueries.join(",")}`);
    check(res, {
      "multiple queries data": (r) => r.status === 200,
    }) || errorRate.add(1);
  } else {
    res = http.get(`${BASE_URL}/api/data`);
    check(res, {
      "all data retrieved": (r) => r.status === 200,
    }) || errorRate.add(1);
  }

  sleep(2);

  res = http.get(`${BASE_URL}/api/data`);
  check(res, {
    "data refresh": (r) => r.status === 200,
  }) || errorRate.add(1);
}

function lightDashboardUser() {
  let res = http.get(`${BASE_URL}/api/auth/status`);
  check(res, {
    "auth status": (r) => r.status === 200 || r.status === 401,
  }) || errorRate.add(1);

  sleep(0.5);

  res = http.get(`${BASE_URL}/api/queries`);
  let queries = [];
  if (res.status === 200 && res.json()) {
    queries = res.json();
  }

  sleep(1);

  if (queries.length > 0) {
    const query = queries[Math.floor(Math.random() * queries.length)];
    res = http.get(`${BASE_URL}/api/data?queries=${query}`);
  } else {
    res = http.get(`${BASE_URL}/api/data`);
  }

  check(res, {
    "single query data": (r) => r.status === 200,
  }) || errorRate.add(1);
}

function queryExplorer() {
  let res = http.get(`${BASE_URL}/api/queries`);
  check(res, {
    "query list": (r) => r.status === 200,
  }) || errorRate.add(1);

  sleep(1);

  const attempts = Math.floor(Math.random() * 3) + 1;
  for (let i = 0; i < attempts; i++) {
    res = http.get(`${BASE_URL}/api/data`);
    check(res, {
      "explore data": (r) => r.status === 200,
    }) || errorRate.add(1);

    sleep(1.5);
  }
}

function healthCheck() {
  const res = http.get(`${BASE_URL}/api/v1/health`);
  check(res, {
    "health check": (r) => r.status === 200 && r.json("status") === "OK",
  }) || errorRate.add(1);
}

export function setup() {
  console.log(`Starting load test against ${BASE_URL}`);

  const res = http.get(`${BASE_URL}/api/v1/health`);
  if (res.status !== 200) {
    throw new Error(`Service not available at ${BASE_URL}`);
  }

  return { baseUrl: BASE_URL };
}

export function teardown(data) {
  console.log("Load test completed");
}
