// Seed E2E scenario — calls POST /e2e/seed/:scenario
// Requires env var: SCENARIO (e.g. "onboarding")
// Optional env var: LOG_PATH — absolute path for backend log capture
const scenario = SCENARIO || 'onboarding';
const body = {};
if (typeof LOG_PATH !== 'undefined' && LOG_PATH) {
    body.log_path = LOG_PATH;
}
const response = http.post('http://localhost:8080/e2e/seed/' + scenario, {
    body: JSON.stringify(body),
    headers: { 'Content-Type': 'application/json' }
});
if (!response.ok) {
    throw new Error('E2E seed failed (status ' + response.status + '): ' + response.body);
}
