/**
 * k6 Load Test: GET /stats/assignments
 * 
 * This script tests the read-only statistics endpoint which returns
 * the number of assignments per user and per PR.
 * 
 * Test Profile:
 * - Ramp up: 0 → 10 VUs over 10 seconds
 * - Steady state: 10 → 50 VUs over 30 seconds  
 * - Ramp down: 50 → 0 VUs over 10 seconds
 * 
 * Usage:
 *   k6 run stats_assignments.js
 *   BASE_URL=http://localhost:8080 k6 run stats_assignments.js
 *   BASE_URL=http://avito-internship-backend-service-test:8080 k6 run stats_assignments.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export const options = {
  stages: [
    { duration: '10s', target: 10 },  // Ramp up to 10 VUs
    { duration: '30s', target: 50 },  // Ramp up to 50 VUs
    { duration: '10s', target: 0 },   // Ramp down to 0 VUs
  ],
  thresholds: {
    http_req_duration: ['p(95)<300'], // 95% of requests should complete within 300ms (SLI requirement)
    http_req_failed: ['rate<0.001'],  // Less than 0.1% failure rate (99.9% SLI requirement)
    errors: ['rate<0.001'],
  },
};

// Get base URL from environment variable or use default
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const url = `${BASE_URL}/stats/assignments`;

  const response = http.get(url);

  // Validate response
  const success = check(response, {
    'status is 200': (r) => r.status === 200,
    'response has body': (r) => r.body && r.body.length > 0,
    'content type is JSON': (r) => r.headers && r.headers['Content-Type'] && r.headers['Content-Type'].includes('application/json'),
  });

  // Track errors
  errorRate.add(!success);

  // Small pause between requests (simulate realistic user behavior)
  sleep(0.1);
}

/**
 * Setup function - runs once at the start
 * Can be used to verify service availability
 */
export function setup() {
  const url = `${BASE_URL}/healthz`;
  const res = http.get(url);
  
  if (res.status !== 200) {
    console.warn(`Warning: Health check failed. Service may not be ready. Status: ${res.status}`);
  } else {
    console.log('✓ Service health check passed');
  }

  return { baseUrl: BASE_URL };
}

/**
 * Teardown function - runs once at the end
 * Can be used for cleanup or final reporting
 */
export function teardown(data) {
  console.log(`Load test completed against ${data.baseUrl}`);
}
