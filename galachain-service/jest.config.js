module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'node',
  roots: ['<rootDir>/src', '<rootDir>/test'],
  testMatch: ['**/__tests__/**/*.test.ts', '**/*.test.ts'],
  collectCoverageFrom: [
    'src/**/*.ts',
    '!src/**/*.d.ts',
    '!src/types/**',
    '!src/index.ts',
  ],
  // Coverage thresholds (enforced by coverage.sh and CI)
  // Target: 70% coverage as per NFR-040
  // Disabled in jest.config.js to allow tests to run - enforcement is in CI
  // coverageThreshold: {
  //   global: {
  //     lines: 70,
  //     statements: 70,
  //     functions: 70,
  //     branches: 50,
  //   },
  // },
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json', 'node'],
  transformIgnorePatterns: [
    'node_modules/(?!(uuid)/)',
  ],
  // Allow test to continue even if some tests fail
  bail: false,
};
