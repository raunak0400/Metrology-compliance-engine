/**
 * Prisma seed — development bootstrap data.
 *
 * Seeds:
 *   - system roles (ADMIN, INSPECTOR, REVIEWER, SUPERVISOR)
 *   - example product categories
 *   - example legal rules + rule requirements (TEMPLATES — see note below)
 *   - a development admin user
 *
 * IMPORTANT: The legal rules seeded here are illustrative EXAMPLES for local
 * development only. They are NOT an authoritative statement of the law. The
 * legal-rule repository must be populated/verified from the authoritative
 * Legal Metrology Act, 2009 and Legal Metrology (Packaged Commodities)
 * Rules, 2011 (and amendments) before any real compliance use.
 *
 * Idempotent: safe to run multiple times (uses upserts on natural keys).
 */
import { PrismaClient } from '@prisma/client';
import argon2 from 'argon2';

const prisma = new PrismaClient();

const SYSTEM_ROLES: { name: string; description: string; permissions: string[] }[] = [
  { name: 'ADMIN', description: 'Full system administrator', permissions: ['*'] },
  {
    name: 'INSPECTOR',
    description: 'Creates inspections, uploads images, triggers processing',
    permissions: [
      'inspections:read',
      'inspections:write',
      'products:read',
      'products:write',
      'scans:write',
      'reports:read',
    ],
  },
  {
    name: 'REVIEWER',
    description: 'Reviews low-confidence / flagged inspections and edits declarations',
    permissions: [
      'inspections:read',
      'declarations:write',
      'violations:write',
      'inspections:finalize',
    ],
  },
  {
    name: 'SUPERVISOR',
    description: 'Oversight, dashboard, rules and audit access',
    permissions: ['inspections:read', 'rules:read', 'rules:write', 'audit:read', 'dashboard:read'],
  },
];

const CATEGORIES: { name: string; code: string; description: string }[] = [
  { name: 'General Packaged Commodity', code: 'GENERAL', description: 'Default catch-all category' },
  { name: 'Packaged Food', code: 'FOOD', description: 'Pre-packaged food commodities' },
  { name: 'Cosmetics & Toiletries', code: 'COSMETICS', description: 'Cosmetic packaged goods' },
  { name: 'Household & Cleaning', code: 'HOUSEHOLD', description: 'Cleaning agents, detergents' },
  { name: 'Consumer Electronics', code: 'ELECTRONICS', description: 'Packaged electronic goods' },
];

// Example rule templates. `requirements` describe how the deterministic engine
// evaluates the rule. Replace/verify against authoritative legal text.
const RULE_TEMPLATES: Array<{
  code: string;
  title: string;
  description: string;
  actReference: string;
  severityDefault: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  requirements: Array<{
    key: string;
    declarationType?: string;
    operator: string;
    expected?: unknown;
    unit?: string;
    message: string;
  }>;
}> = [
  {
    code: 'PCR2011-MRP-PRESENCE',
    title: 'Retail sale price (MRP) declaration present',
    description: 'Every package must declare the retail sale price (MRP inclusive of all taxes).',
    actReference: 'LM (Packaged Commodities) Rules, 2011 — Rule 6 / 18 (EXAMPLE)',
    severityDefault: 'HIGH',
    requirements: [
      {
        key: 'declaration.MRP.present',
        declarationType: 'MRP',
        operator: 'EXISTS',
        message: 'MRP declaration is missing on the package.',
      },
      {
        key: 'declaration.MRP.format',
        declarationType: 'MRP',
        operator: 'REGEX_MATCH',
        expected: { pattern: '(?i)(m\\.?r\\.?p\\.?|maximum retail price).*?\\d' },
        message: 'MRP is present but not in an expected "MRP Rs. <amount>" format.',
      },
    ],
  },
  {
    code: 'PCR2011-NETQTY-PRESENCE',
    title: 'Net quantity declaration present',
    description: 'Every package must declare the net quantity in standard units.',
    actReference: 'LM (Packaged Commodities) Rules, 2011 — Rule 6 (EXAMPLE)',
    severityDefault: 'HIGH',
    requirements: [
      {
        key: 'declaration.NET_QUANTITY.present',
        declarationType: 'NET_QUANTITY',
        operator: 'EXISTS',
        message: 'Net quantity declaration is missing.',
      },
    ],
  },
  {
    code: 'PCR2011-MANUFACTURER-DETAILS',
    title: 'Manufacturer / packer / importer name and address',
    description: 'The name and complete address of the manufacturer, packer or importer must be declared.',
    actReference: 'LM (Packaged Commodities) Rules, 2011 — Rule 6 (EXAMPLE)',
    severityDefault: 'MEDIUM',
    requirements: [
      {
        key: 'declaration.MANUFACTURER.present',
        declarationType: 'MANUFACTURER',
        operator: 'NOT_EMPTY',
        message: 'Manufacturer/packer/importer name and address not found.',
      },
    ],
  },
  {
    code: 'PCR2011-CONSUMER-CARE',
    title: 'Consumer care details present',
    description: 'Consumer care contact details (name, address, phone/email) must be declared.',
    actReference: 'LM (Packaged Commodities) Rules, 2011 — Rule 6 (EXAMPLE)',
    severityDefault: 'MEDIUM',
    requirements: [
      {
        key: 'declaration.CONSUMER_CARE.present',
        declarationType: 'CONSUMER_CARE',
        operator: 'NOT_EMPTY',
        message: 'Consumer care details not found.',
      },
    ],
  },
  {
    code: 'PCR2011-COUNTRY-OF-ORIGIN',
    title: 'Country of origin (imported packages)',
    description: 'Imported packages must declare the country of origin.',
    actReference: 'LM (Packaged Commodities) Rules, 2011 — Rule 6 (EXAMPLE)',
    severityDefault: 'MEDIUM',
    requirements: [
      {
        key: 'declaration.COUNTRY_OF_ORIGIN.requiredIfImported',
        declarationType: 'COUNTRY_OF_ORIGIN',
        operator: 'REQUIRED_IF_IMPORTED',
        message: 'Country of origin not declared for an imported package.',
      },
    ],
  },
  {
    code: 'PCR2011-MFG-DATE',
    title: 'Date of manufacture / packing / import',
    description: 'The month and year of manufacture, pre-packing or import must be declared.',
    actReference: 'LM (Packaged Commodities) Rules, 2011 — Rule 6 (EXAMPLE)',
    severityDefault: 'MEDIUM',
    requirements: [
      {
        key: 'declaration.MANUFACTURE_DATE.valid',
        declarationType: 'MANUFACTURE_DATE',
        operator: 'DATE_VALID',
        message: 'Date of manufacture/packing missing or not a valid date.',
      },
    ],
  },
  {
    code: 'PCR2011-MIN-TEXT-HEIGHT',
    title: 'Minimum height of declarations (measurement)',
    description:
      'Mandatory declarations must meet the minimum letter/numeral height based on the principal display area. Requires a valid scale reference; otherwise flagged for manual review.',
    actReference: 'LM (Packaged Commodities) Rules, 2011 — Rule 8 (EXAMPLE)',
    severityDefault: 'LOW',
    requirements: [
      {
        key: 'measurement.declaration.minHeight',
        declarationType: 'NET_QUANTITY',
        operator: 'MIN_TEXT_HEIGHT_MM',
        expected: { minMm: 1.0 },
        unit: 'mm',
        message: 'Declaration text height appears below the minimum required height.',
      },
    ],
  },
];

async function seedRoles() {
  for (const role of SYSTEM_ROLES) {
    await prisma.role.upsert({
      where: { name: role.name },
      update: { description: role.description, permissions: role.permissions, isSystem: true },
      create: { name: role.name, description: role.description, permissions: role.permissions, isSystem: true },
    });
  }
  console.log(`Seeded ${SYSTEM_ROLES.length} roles.`);
}

async function seedCategories() {
  for (const c of CATEGORIES) {
    await prisma.productCategory.upsert({
      where: { code: c.code },
      update: { name: c.name, description: c.description },
      create: { name: c.name, code: c.code, description: c.description },
    });
  }
  console.log(`Seeded ${CATEGORIES.length} product categories.`);
}

async function seedRules() {
  const general = await prisma.productCategory.findUnique({ where: { code: 'GENERAL' } });
  for (const t of RULE_TEMPLATES) {
    const rule = await prisma.legalRule.upsert({
      where: { code: t.code },
      update: {
        title: t.title,
        description: t.description,
        actReference: t.actReference,
        severityDefault: t.severityDefault as never,
        isActive: true,
      },
      create: {
        code: t.code,
        title: t.title,
        description: t.description,
        actReference: t.actReference,
        severityDefault: t.severityDefault as never,
        isActive: true,
      },
    });

    // Replace requirements deterministically for idempotency.
    await prisma.ruleRequirement.deleteMany({ where: { ruleId: rule.id } });
    await prisma.ruleRequirement.createMany({
      data: t.requirements.map((r, i) => ({
        ruleId: rule.id,
        key: r.key,
        declarationType: (r.declarationType ?? null) as never,
        operator: r.operator as never,
        expected: (r.expected ?? undefined) as never,
        unit: r.unit ?? null,
        message: r.message,
        orderIndex: i,
      })),
    });

    // Map every example rule to the GENERAL category.
    if (general) {
      await prisma.ruleCategory.upsert({
        where: { ruleId_categoryId: { ruleId: rule.id, categoryId: general.id } },
        update: {},
        create: { ruleId: rule.id, categoryId: general.id },
      });
    }
  }
  console.log(`Seeded ${RULE_TEMPLATES.length} example legal rules with requirements.`);
}

async function seedAdmin() {
  const email = process.env.ADMIN_EMAIL ?? 'admin@legalmetrology.local';
  const password = process.env.ADMIN_PASSWORD ?? 'ChangeMe123!';
  const passwordHash = await argon2.hash(password);

  const admin = await prisma.user.upsert({
    where: { email },
    update: { isActive: true },
    create: { email, passwordHash, fullName: 'Development Admin', isActive: true },
  });

  const adminRole = await prisma.role.findUnique({ where: { name: 'ADMIN' } });
  if (adminRole) {
    await prisma.userRole.upsert({
      where: { userId_roleId: { userId: admin.id, roleId: adminRole.id } },
      update: {},
      create: { userId: admin.id, roleId: adminRole.id },
    });
  }
  console.log(`Seeded admin user: ${email} (password: ${password})`);
}

async function main() {
  console.log('Seeding database...');
  await seedRoles();
  await seedCategories();
  await seedRules();
  await seedAdmin();
  console.log('Seeding complete.');
}

main()
  .catch((e) => {
    console.error(e);
    process.exit(1);
  })
  .finally(async () => {
    await prisma.$disconnect();
  });
