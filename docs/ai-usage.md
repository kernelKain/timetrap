# AI usage

OpenAI Codex assisted throughout TimeTrap development with:

- requirements decomposition and implementation planning;
- Go domain, analyzer, transport, and persistence implementation;
- React component, accessibility, and responsive design work;
- focused regression, handler, integration, and frontend tests;
- Zerops build and runtime configuration;
- documentation structure and technical copy;
- live deployment inspection and verification through ZCP.

AI output was not treated as evidence by itself. Changes were reviewed against repository contracts, formatted and compiled, tested with the repository suites, built as production artifacts, and checked against the live API and browser. Database operations used explicit migration and test-database safeguards, and secrets were kept out of prompts, source, output, and documentation.

AI assistance can still miss accessibility nuances, unusual browser behavior, incomplete models, and operational conditions not represented in tests. The deterministic analyzer and server responses—not generated prose—remain authoritative for TimeTrap results.
