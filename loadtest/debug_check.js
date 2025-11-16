/**
 * Debug script to check if seed data exists
 */

import http from 'k6/http';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  vus: 1,
  iterations: 1,
};

export default function () {
  console.log('Checking if test users exist...\n');

  const users = [
    'load-test-user-1',
    'load-test-user-2',
    'load-test-user-3',
    'load-test-user-4',
    'load-test-user-5',
    'load-test-user-6',
    'load-test-user-7',
    'load-test-user-8',
  ];

  users.forEach(userId => {
    const res = http.get(`${BASE_URL}/users/getReview?user_id=${userId}`);
    console.log(`User ${userId}: status=${res.status}`);
    if (res.status !== 200) {
      console.log(`  Response: ${res.body}`);
    }
  });

  console.log('\nChecking teams...');
  const teams = ['load-test-team-1', 'load-test-team-2'];
  teams.forEach(teamName => {
    const res = http.get(`${BASE_URL}/team/get?team_name=${teamName}`);
    console.log(`Team ${teamName}: status=${res.status}`);
    if (res.status !== 200) {
      console.log(`  Response: ${res.body}`);
    }
  });

  console.log('\nTrying to create a test PR...');
  const prPayload = JSON.stringify({
    pull_request_id: 'debug-pr-test',
    pull_request_name: 'Debug PR',
    author_id: 'load-test-user-1',
  });
  
  const prRes = http.post(
    `${BASE_URL}/pullRequest/create`,
    prPayload,
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  console.log(`\nPR Creation: status=${prRes.status}`);
  console.log(`Response: ${prRes.body}`);
}
