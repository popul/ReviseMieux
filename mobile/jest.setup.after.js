// Runs AFTER the test framework is initialized (setupFilesAfterEnv).
// Silences Animated(View) act() warnings from MasteryBar animations.
// These warnings are a known React Native testing issue — the animations
// are asynchronous and fire outside act() boundaries.

const originalError = console.error;
console.error = (...args) => {
  if (typeof args[0] === 'string' && args[0].includes('not wrapped in act')) {
    return; // suppress act() warnings from Animated
  }
  originalError.call(console, ...args);
};
