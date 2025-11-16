/**
 * k6 Load Test: POST /pullRequest/create
 * 
 * This script tests the write-heavy PR creation endpoint which automatically
 * assigns up to 2 reviewers from the author's team.
 * 
 * Test Profile:
 * - Ramp up: 0 → 5 VUs over 10 seconds
 * - Steady state: 5 → 20 VUs over 30 seconds
 * - Ramp down: 20 → 0 VUs over 10 seconds
 * 
 * Prerequisites:
 * - Database must be seeded with teams and users
 * - Run seed_data.js before this test
 * 
 * Usage:
 *   k6 run create_pr.js
 *   BASE_URL=http://localhost:8080 k6 run create_pr.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const prsCreated = new Counter('prs_created');
const prsDuplicate = new Counter('prs_duplicate');

// Test configuration
export const options = {
  stages: [
    { duration: '10s', target: 5 },   // Ramp up to 5 VUs
    { duration: '30s', target: 20 },  // Ramp up to 20 VUs
    { duration: '10s', target: 0 },   // Ramp down to 0 VUs
  ],
  thresholds: {
    http_req_duration: ['p(95)<300'], // 95% of requests should complete within 300ms
    http_req_failed: ['rate<0.05'],   // Allow up to 5% failure (due to conflicts/duplicates)
    errors: ['rate<0.05'],
  },
};

// Get base URL from environment variable or use default
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Test data - authors from seeded teams
// These should match the users created by seed_data.js
const AUTHORS = [
  'load-test-user-1',
  'load-test-user-2',
  'load-test-user-3',
  'load-test-user-4',
  'load-test-user-5',
  'load-test-user-6',
  'load-test-user-7',
  'load-test-user-8',
];

let requestCounter = 0;

export default function () {
  const url = `${BASE_URL}/pullRequest/create`;

  // Generate unique PR ID using timestamp, VU ID, and counter
  const vuId = __VU;
  const timestamp = Date.now();
  requestCounter++;
  const uniqueId = `pr-loadtest-${vuId}-${timestamp}-${requestCounter}`;

  // Select random author from test users
  const authorId = AUTHORS[Math.floor(Math.random() * AUTHORS.length)];

  const payload = JSON.stringify({
    pull_request_id: uniqueId,
    pull_request_name: `Load Test PR ${requestCounter}`,
    author_id: authorId,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const response = http.post(url, payload, params);

  // Validate response
  const success = check(response, {
    'status is 201 or 409': (r) => r.status === 201 || r.status === 409,
    'status is 201 (created)': (r) => r.status === 201,
    'response has body': (r) => r.body && r.body.length > 0,
    'content type is JSON': (r) => r.headers && r.headers['Content-Type'] && r.headers['Content-Type'].includes('application/json'),
  });

  // Track specific outcomes
  if (response.status === 201) {
    prsCreated.add(1);
  } else if (response.status === 409) {
    prsDuplicate.add(1);
  }

  // Track errors (excluding expected 409 conflicts)
  const isError = !success || (response.status !== 201 && response.status !== 409);
  errorRate.add(isError);

  // Log unexpected errors
  if (response.status !== 201 && response.status !== 409) {
    console.error(`Unexpected status ${response.status} for PR ${uniqueId}: ${response.body}`);
  }

  // Small pause between requests
  sleep(0.2);
}

/**
 * Setup function - verify service and seed data availability
 */
export function setup() {
  console.log('Verifying service availability...');
  
  const healthUrl = `${BASE_URL}/healthz`;
  const healthRes = http.get(healthUrl);
  
  if (healthRes.status !== 200) {
    throw new Error(`Service health check failed. Status: ${healthRes.status}`);
  }
  
  console.log('✓ Service health check passed');

  // Verify at least one test user exists
  const testUrl = `${BASE_URL}/users/getReview?user_id=${AUTHORS[0]}`;
  const testRes = http.get(testUrl);
  
  if (testRes.status === 404) {
    console.warn('⚠ Test users not found. Please run seed_data.js first!');
    console.warn('  Command: k6 run seed_data.js');
    throw new Error('Test data not seeded. Run seed_data.js before this test.');
  }

  console.log('✓ Test data verification passed');
  console.log(`Starting load test with ${AUTHORS.length} test authors`);

  return { baseUrl: BASE_URL, authors: AUTHORS };
}

/**
 * Teardown function - summary reporting
 */
export function teardown(data) {
  console.log(`\nLoad test completed against ${data.baseUrl}`);
  console.log(`Test used ${data.authors.length} different author accounts`);
}
