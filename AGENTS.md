# AGENTS.md

## Project Direction

Yuyan（语燕） is an AI pet app, not a Phase 1 instant messaging app.
The Phase 1 product loop is: account login -> create pet -> text conversation -> memory -> growth -> entitlement and analytics foundation.

## Required Reading

Before making changes, read:

- `docs/LLM_DEV_GUIDE.md`
- `docs/MILESTONES.md`
- The relevant OpenSpec change under `openspec/changes/`

## Hard Rules

- Do not expand friend, contact, or human-to-human IM features unless explicitly requested.
- Do not rewrite or delete historical migrations.
- Do not let business services call model provider SDKs directly; use AI Gateway.
- Do not use offset pagination for messages.
- Do not let the model decide core growth values, entitlement, or billing state.
- Add or update tests for changed behavior.
- If changing API, database, or protocol behavior, create or update an OpenSpec change first.

## Stack

- Server: Go monolith, Gin, GORM, MySQL 8, Redis.
- App: Flutter, Riverpod, drift, Dio, WebSocket.
- Deployment: Docker Compose for Phase 1.
