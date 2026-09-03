# UBOM Initial Release Plan

## Objective

Ship a small, durable part log. A user can define the language for part
numbers, organize that language into a taxonomy, and create searchable parts
against the taxonomy.

The release should make the sequence definition and taxonomy the source of
truth for part creation without expanding into BOM, artifact, or quality-system
workflows.

## Release Scope

### Included

- User-created sequence definitions.
- User-created taxonomy trees bound to sequence definitions.
- Taxonomy predicates and inherited/overridden attribute definitions.
- Creation of part numbers against a selected taxonomy node.
- Entry and validation of part attribute values during part creation.
- A searchable part log.
- A home page with part-number search, summary information, and taxonomy-driven
  drill-down.
- Durable Postgres persistence.
- Sequence and taxonomy revision history sufficient to preserve the definition
  used when a part was created.
- A minimal Docker Compose deployment for the API, frontend, and Postgres.

### Explicitly Deferred

- BOM authoring and multi-BOM support.
- Artifact sources, snapshots, and file integrations.
- Release approval workflows and formal ECO/ECR change control.
- CAD integrations.
- Advanced reporting and analytics.
- Fine-grained permissions beyond the initial deployment boundary.

## Persistence Strategy

Use Postgres for deployed environments. Keep SQLite and the in-memory store for
local development and tests where they provide fast, isolated feedback.

The existing `store.Store` interface remains the domain boundary. Add a
separate `PostgresStore` rather than making the SQLite SQL dialect conditional.
Both stores must pass the shared store contract tests.

### Required persistence work

- Add the Postgres driver and connection-pool setup.
- Add versioned, repeatable database migrations.
- Add `PostgresStore` implementations for sequence definitions, taxonomies,
  attributes, part drafts, parts, and revisions.
- Preserve uniqueness constraints for definition IDs, revision IDs, and part
  numbers.
- Use transactions for taxonomy/sequence revision creation and part creation.
- Add integration tests against a real Postgres instance.
- Add DSN-based configuration, such as `UBOM_DATABASE_URL`.
- Add health and readiness checks that include database connectivity.

The current Compose file starts Postgres, but UBOM still uses SQLite until this
adapter and configuration work is complete.

## Domain Decisions

The initial release adopts these rules:

- Part-number values are globally unique canonical values across the application.
- A committed part retains the exact sequence and taxonomy revisions used to
  generate and classify it.
- Sequence and taxonomy revisions are immutable. An edit creates a new revision
  and affects future part creation only; existing parts do not move
  automatically.
- A part must resolve to exactly one leaf or explicitly part-bearing taxonomy
  node. A root-only or otherwise incomplete classification is not sufficient.
- A selected taxonomy context searches its descendants, but classification must
  produce exactly one final match.
- No matching taxonomy branch is a classification error.
- Multiple matching branches are an ambiguity error; there is no implicit
  precedence rule in the initial release.
- Drafts are temporary. A draft does not allocate or reserve a new part number;
  allocation occurs transactionally only when the draft is released.
- A released draft cannot be edited or released again. An abandoned draft
  cannot be edited or released.
- The initial release focuses on part numbers. Part revisions and BOM behavior
  are deferred and are not required to create a part number.
- Committed parts are never silently deleted. An archive/status field provides
  tombstone behavior without removing the historical record.
- Attribute definitions are strongly typed enough to validate string, decimal,
  enum, and required values before commit.
- The initial deployment is explicitly single-user; it does not claim
  multi-user ownership or authorization.

## User Workflows

### Sequence definition

- Create a named sequence definition.
- Build and validate its grammar.
- Preview parsing and generation.
- Save immutable revisions.
- Reopen prior revisions read-only.

### Taxonomy

- Create a taxonomy for a sequence definition.
- Add, rename, reorder, and nest categories.
- Add and edit predicates using known grammar fields.
- Define attributes at any node.
- Verify inherited and overridden attributes.
- Save immutable taxonomy revisions.
- Prevent invalid or ambiguous taxonomy definitions from being saved.

### Part creation

- Start from a taxonomy node or the home page. A taxonomy drill-down can create
  a context-sensitive part using the current node as its starting context.
- Progressively disclose the remaining inputs as the user enters text or
  selects a taxonomy path.
- Enter known sequence bindings.
- Enter required and optional attribute values.
- Preview the generated part number.
- Validate classification and attributes server-side.
- Allocate and commit the part number only on release, with its definition
  revision references.
- Show the resulting part summary.

### Part log and navigation

- Search by complete or partial part number.
- Filter by taxonomy path.
- Filter by attribute values.
- Drill down from taxonomy category to child category to matching parts.
- Show part number, taxonomy path, definition revisions, attributes, and status.
- Support pagination, sorting, loading, empty, and error states.

## Frontend and Deployment

- Replace the current development-only Vite proxy arrangement with a deployable
  frontend container or API-served static assets.
- Configure the frontend API origin for local Compose and deployed environments.
- Add an API container healthcheck and a Postgres readiness dependency.
- Store credentials through environment variables or deployment secrets, never
  committed Compose defaults outside local development.
- Define persistent volumes and a backup/restore procedure for Postgres.
- Add a production-like Compose smoke test that creates a sequence, taxonomy,
  and part through the public UI/API.

## Security and Operations

The first release uses an explicitly single-user deployment-level access
boundary. It must not imply multi-user authorization or ownership that does not
exist.

Before release, add:

- Document the single-user deployment assumption.
- Request and error logging without leaking credentials.
- Consistent API error responses.
- Database migration and rollback procedures.
- Audit fields for created/updated timestamps and the actor or deployment
  identity available at the time.
- Backup, restore, and data-loss expectations.

## Implementation Order

1. Finalize the domain decisions above and lock API contracts for definitions,
   drafts, committed parts, search, and summaries.
2. Implement Postgres migrations and `PostgresStore`; run the store contract
   suite against SQLite and Postgres.
3. Add DSN configuration, connection lifecycle, health/readiness endpoints, and
   real-Postgres Compose startup.
4. Complete server-side part creation, attribute-value validation, uniqueness,
   and revision references.
5. Build the home page, search, part summary, taxonomy drill-down, pagination,
   and responsive states.
6. Package the frontend and API together in Compose and run an end-to-end smoke
   test against Postgres.
7. Add deployment documentation, backup instructions, and release QA coverage.

## Release Acceptance

The release is ready when a clean deployment can complete this path without
manual database work:

1. Start the API, frontend, and Postgres with Docker Compose.
2. Create a sequence definition and save a revision.
3. Create a taxonomy with nested categories, predicates, and attributes.
4. Create a part with valid bindings and required attribute values.
5. Reject invalid bindings, invalid attributes, duplicate part numbers, and
   ambiguous taxonomy matches with actionable errors.
6. Find the committed part from the home-page search.
7. Reach the part through taxonomy drill-down.
8. Reload the deployment and verify the definitions, taxonomy, part, and
   revision references persist.
9. Open historical definitions read-only.
10. Confirm the database can be backed up and restored using the documented
    procedure.
