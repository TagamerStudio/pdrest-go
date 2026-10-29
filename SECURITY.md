# Security Policy

## Supported Versions

Only the latest release line receives security fixes.

| Version | Supported          |
| ------- | ------------------ |
| 0.3.x   | :white_check_mark: |
| < 0.3   | :x:                |

## Reporting a Vulnerability

Please do not report security vulnerabilities in public issues.

Use GitHub's private vulnerability reporting form:

<https://github.com/TagamerStudio/pdrest-go/security/advisories/new>

If the form is unavailable, open a minimal public issue asking a maintainer for
a private contact channel. Do not include any vulnerability details in that
issue.

Include, when possible:

- the affected version of this library;
- the Go version and operating system;
- the PalDefender and Palworld server versions;
- a minimal reproduction;
- the potential impact;
- relevant logs with credentials and personal data removed.

Never include PalDefender credentials, API tokens, server addresses, player
data or other secrets in a report.

## Response

Reports are handled on a best-effort basis. We aim to acknowledge valid reports
within 3 business days and to coordinate disclosure after a fix is available.
Please allow time for a release before disclosing the issue publicly.

## Scope

This policy covers vulnerabilities in the pdrest-go library, including base URL
handling, bearer-token handling, TLS behavior, response size limits, redirect
handling and request/response validation.

The following are out of scope:

- vulnerabilities in PalDefender or the Palworld dedicated server; report those
  to their maintainers;
- the lack of transport encryption when the PalDefender REST API is served
  directly over plain HTTP. The API sends a bearer token in the `Authorization`
  header, so anyone who can observe the traffic can read the token. Run it on a
  trusted LAN or behind a TLS-terminating reverse proxy and pass an `https://`
  base URL to `NewClient`.

## Secure Usage

- Keep the PalDefender REST API on a trusted network or behind a TLS reverse
  proxy.
- Surrounding whitespace is trimmed from the token; the remaining value is sent
  as provided. Avoid embedding credentials or other secrets in the base URL.
- Treat `APIError.ResponseBody` as diagnostic data and avoid logging it if it
  may contain sensitive information.
- When injecting an HTTP client with `WithHTTPClient`, configure its timeout,
  redirect, proxy and TLS policies safely.
