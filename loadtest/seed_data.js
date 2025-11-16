/**
 * k6 Seed Data Script
 * 
 * This script populates the database with test teams and users required
 * for running load tests. It should be executed once before running the
 * main load test scripts.
 * 
 * Creates:
 * - 2 test teams (load-test-team-1, load-test-team-2)
 * - 8 test users distributed across teams
 * 
 * Usage:
 *   k6 run seed_data.js
 *   BASE_URL=http://localhost:8080 k6 run seed_data.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';

// Get base URL from environment variable or use default
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  vus: 1,
  iterations: 1,
};

export default function () {
  console.log('Starting database seeding for load tests...\n');

  const headers = {
    'Content-Type': 'application/json',
  };

  // Team 1: load-test-team-1 with 5 members
  console.log('Creating team: load-test-team-1...');
  const team1 = {
    team_name: 'load-test-team-1',
    members: [
      { user_id: 'load-test-user-1', username: 'LoadTest User 1', is_active: true },
      { user_id: 'load-test-user-2', username: 'LoadTest User 2', is_active: true },
      { user_id: 'load-test-user-3', username: 'LoadTest User 3', is_active: true },
      { user_id: 'load-test-user-4', username: 'LoadTest User 4', is_active: true },
      { user_id: 'load-test-user-5', username: 'LoadTest User 5', is_active: true },
    ],
  };

  const res1 = http.post(
    `${BASE_URL}/team/add`,
    JSON.stringify(team1),
    { headers }
  );

  const success1 = check(res1, {
    'team-1 created (201) or already exists (400)': (r) => r.status === 201 || r.status === 400,
  });

  if (res1.status === 201) {
    console.log('✓ Team load-test-team-1 created successfully');
  } else if (res1.status === 400) {
    console.log('ℹ Team load-test-team-1 already exists');
  } else {
    console.error(`✗ Failed to create team-1: ${res1.status} - ${res1.body}`);
  }

  sleep(0.5);

  // Team 2: load-test-team-2 with 3 members
  console.log('Creating team: load-test-team-2...');
  const team2 = {
    team_name: 'load-test-team-2',
    members: [
      { user_id: 'load-test-user-6', username: 'LoadTest User 6', is_active: true },
      { user_id: 'load-test-user-7', username: 'LoadTest User 7', is_active: true },
      { user_id: 'load-test-user-8', username: 'LoadTest User 8', is_active: true },
    ],
  };

  const res2 = http.post(
    `${BASE_URL}/team/add`,
    JSON.stringify(team2),
    { headers }
  );

  const success2 = check(res2, {
    'team-2 created (201) or already exists (400)': (r) => r.status === 201 || r.status === 400,
  });

  if (res2.status === 201) {
    console.log('✓ Team load-test-team-2 created successfully');
  } else if (res2.status === 400) {
    console.log('ℹ Team load-test-team-2 already exists');
  } else {
    console.error(`✗ Failed to create team-2: ${res2.status} - ${res2.body}`);
  }

  sleep(0.5);

  // Verify teams were created
  console.log('\nVerifying created teams...');
  
  const verifyRes1 = http.get(`${BASE_URL}/team/get?team_name=load-test-team-1`);
  const verify1 = check(verifyRes1, {
    'team-1 exists': (r) => r.status === 200,
  });

  const verifyRes2 = http.get(`${BASE_URL}/team/get?team_name=load-test-team-2`);
  const verify2 = check(verifyRes2, {
    'team-2 exists': (r) => r.status === 200,
  });

  console.log('\n========================================');
  console.log('Seeding Summary:');
  console.log('========================================');
  console.log(`Team 1 (load-test-team-1): ${verify1 ? '✓' : '✗'} (5 members)`);
  console.log(`Team 2 (load-test-team-2): ${verify2 ? '✓' : '✗'} (3 members)`);
  console.log('Total users: 8');
  console.log('========================================\n');

  if (verify1 && verify2) {
    console.log('✓ Database seeding completed successfully!');
    console.log('You can now run the load test scripts:\n');
    console.log('  k6 run stats_assignments.js');
    console.log('  k6 run create_pr.js\n');
  } else {
    console.error('✗ Seeding verification failed. Please check the errors above.');
  }
}

/**
 * Setup function - verify service availability
 */
export function setup() {
  console.log('Checking service health...');
  
  const healthUrl = `${BASE_URL}/healthz`;
  const healthRes = http.get(healthUrl);
  
  if (healthRes.status !== 200) {
    throw new Error(`Service is not available. Health check failed with status: ${healthRes.status}`);
  }
  
  console.log('✓ Service is healthy and ready\n');
  
  return { baseUrl: BASE_URL };
}
