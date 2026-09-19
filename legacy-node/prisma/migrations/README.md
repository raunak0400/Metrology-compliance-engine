# Prisma migrations

This folder holds generated SQL migrations. It is intentionally empty in the
skeleton because migrations must be generated against a running PostgreSQL
instance.

## Generate the initial migration

With PostgreSQL running and `DATABASE_URL` set in `.env`:

```bash
npm run prisma:generate      # generate the typed Prisma Client
npm run prisma:migrate       # prisma migrate dev --name init  (creates SQL here)
npm run db:seed              # populate roles, categories, example rules, admin
```

For production / CI, apply committed migrations without prompting:

```bash
npm run prisma:deploy        # prisma migrate deploy
```

Commit the generated `migrations/<timestamp>_init/` directory so the schema is
reproducible across environments.
