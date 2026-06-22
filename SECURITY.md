# Security Policy

## Reporting a vulnerability

Please report security issues **privately** — do not open a public issue.

Use GitHub's private vulnerability reporting:
[**Report a vulnerability**](https://github.com/NerdMeNot/doze/security/advisories/new)
(Security → Advisories → Report a vulnerability on the repo).

Please include a description, reproduction steps, affected version
(`doze version`), and impact. We'll acknowledge your report and work with you on a
fix and disclosure timeline.

## Scope

doze is a **local development** tool: it runs database engines on `127.0.0.1` and
injects dummy credentials for the built-in AWS services. It is not designed to be
exposed to a network or used in production. Reports most relevant to doze include
local privilege escalation, arbitrary code execution via config, or supply-chain
issues in the binary download/verification path.
