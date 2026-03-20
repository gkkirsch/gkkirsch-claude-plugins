---
name: graphql-codegen
description: >
  GraphQL code generation with graphql-codegen — type-safe operations,
  typed document nodes, React hooks generation, schema types, and CI integration.
  Triggers: "graphql codegen", "graphql types", "graphql typescript",
  "typed graphql", "graphql-codegen", "codegen graphql".
  NOT for: Apollo Server/Client setup (use apollo-development).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# GraphQL Code Generation

## Setup

```bash
npm install -D @graphql-codegen/cli @graphql-codegen/typescript \
  @graphql-codegen/typescript-operations @graphql-codegen/typescript-react-apollo \
  @graphql-codegen/client-preset
```

## Configuration

```typescript
// codegen.ts
import type { CodegenConfig } from "@graphql-codegen/cli";

const config: CodegenConfig = {
  // Schema source (pick one)
  schema: "http://localhost:4000/graphql",        // Running server
  // schema: "./schema.graphql",                  // Local SDL file
  // schema: "./src/graphql/schema/**/*.graphql",  // Multiple files

  // Where to find operations (queries, mutations, subscriptions)
  documents: ["src/**/*.{ts,tsx}", "!src/gql/**/*"],

  // Ignore generated files
  ignoreNoDocuments: true,

  generates: {
    // Client preset (recommended for Apollo Client 3+)
    "./src/gql/": {
      preset: "client",
      presetConfig: {
        gqlTagName: "gql",
        fragmentMasking: { unmaskFunctionName: "getFragmentData" },
      },
      config: {
        scalars: {
          DateTime: "string",
          JSON: "Record<string, unknown>",
          Upload: "File",
        },
        enumsAsTypes: true,             // Use union types instead of enums
        skipTypename: false,            // Keep __typename for cache
        dedupeFragments: true,
        nonOptionalTypename: true,
      },
    },

    // Server-side types (for resolvers)
    "./src/types/graphql.ts": {
      plugins: [
        "typescript",
        "typescript-resolvers",
      ],
      config: {
        useIndexSignature: true,
        contextType: "../context#Context",   // Your context type
        mapperTypeSuffix: "Model",
        mappers: {
          User: "../models/user#UserModel",   // Map GraphQL types to DB models
          Post: "../models/post#PostModel",
        },
      },
    },
  },
};

export default config;
```

## Running Codegen

```bash
# Generate once
npx graphql-codegen

# Watch mode (regenerate on file changes)
npx graphql-codegen --watch

# Add to package.json
{
  "scripts": {
    "codegen": "graphql-codegen",
    "codegen:watch": "graphql-codegen --watch"
  }
}
```

## Client Preset Usage (Recommended)

The `client` preset generates typed document nodes that work with Apollo Client's `useQuery`, `useMutation`, etc.

```tsx
// src/components/UserList.tsx
import { useQuery } from "@apollo/client";
import { gql } from "../gql";  // Generated gql function (NOT from @apollo/client)

// The gql function auto-types this query
const GET_USERS = gql(`
  query GetUsers($first: Int!, $after: String) {
    users(first: $first, after: $after) {
      edges {
        cursor
        node {
          id
          name
          email
          ...UserAvatar
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`);

function UserList() {
  // data is fully typed — data.users.edges[0].node.name is string
  // variables are typed — { first: number, after?: string }
  const { data, loading, error } = useQuery(GET_USERS, {
    variables: { first: 20 },
  });

  if (loading) return <p>Loading...</p>;
  if (error) return <p>Error: {error.message}</p>;

  return (
    <ul>
      {data?.users.edges.map(({ node }) => (
        <li key={node.id}>{node.name} — {node.email}</li>
      ))}
    </ul>
  );
}
```

## Fragment Colocation

```tsx
// src/components/UserAvatar.tsx
import { gql } from "../gql";
import { getFragmentData } from "../gql";
import type { FragmentType } from "../gql";

// Define fragment for this component's data needs
const USER_AVATAR_FRAGMENT = gql(`
  fragment UserAvatar on User {
    id
    name
    avatarUrl
  }
`);

// Accept fragment type as prop
function UserAvatar({ user }: { user: FragmentType<typeof USER_AVATAR_FRAGMENT> }) {
  // Unmask the fragment data
  const userData = getFragmentData(USER_AVATAR_FRAGMENT, user);

  return (
    <img
      src={userData.avatarUrl || "/default-avatar.png"}
      alt={userData.name}
    />
  );
}
```

```tsx
// src/components/UserCard.tsx — compose fragments
import { gql } from "../gql";

const GET_USER = gql(`
  query GetUser($id: ID!) {
    user(id: $id) {
      id
      name
      email
      createdAt
      ...UserAvatar
    }
  }
`);

function UserCard({ userId }: { userId: string }) {
  const { data } = useQuery(GET_USER, { variables: { id: userId } });

  return (
    <div>
      {/* Pass the whole user object — fragment masking handles the rest */}
      <UserAvatar user={data!.user!} />
      <h2>{data?.user?.name}</h2>
      <p>{data?.user?.email}</p>
    </div>
  );
}
```

## Typed Mutations

```tsx
import { useMutation } from "@apollo/client";
import { gql } from "../gql";

const CREATE_USER = gql(`
  mutation CreateUser($input: CreateUserInput!) {
    createUser(input: $input) {
      user {
        id
        name
        email
      }
      errors {
        field
        message
      }
    }
  }
`);

function CreateUserForm() {
  // Variables and return type are fully typed
  const [createUser, { loading }] = useMutation(CREATE_USER);

  const handleSubmit = async (formData: { name: string; email: string }) => {
    const { data } = await createUser({
      variables: { input: formData },
      // TypeScript knows the shape of optimisticResponse
      optimisticResponse: {
        createUser: {
          __typename: "CreateUserPayload",
          user: {
            __typename: "User",
            id: "temp",
            name: formData.name,
            email: formData.email,
          },
          errors: [],
        },
      },
    });

    if (data?.createUser.errors.length) {
      // errors is typed as { field: string; message: string }[]
      return data.createUser.errors;
    }
  };
}
```

## Server-Side Types (Resolver Types)

```typescript
// Generated at ./src/types/graphql.ts
// These types are used in resolver implementations

import { Resolvers } from "./types/graphql";

const resolvers: Resolvers = {
  Query: {
    // TypeScript enforces correct return types
    user: async (_, { id }, context) => {
      // Return type must match UserModel (from mapper config)
      return context.db.users.findUnique({ where: { id } });
    },
    users: async (_, { first, after }, context) => {
      // Return type must match UserConnection
      return context.db.users.paginate({ first, after });
    },
  },
  // TypeScript enforces all required resolvers are implemented
  Mutation: {
    createUser: async (_, { input }, context) => {
      const user = await context.db.users.create({ data: input });
      return { user, errors: [] };
    },
  },
};
```

## Custom Scalars

```typescript
// codegen.ts — map GraphQL scalars to TypeScript types
config: {
  scalars: {
    DateTime: "string",                    // ISO 8601 string
    Date: "string",                        // YYYY-MM-DD
    JSON: "Record<string, unknown>",       // Arbitrary JSON
    BigInt: "bigint",                       // Large integers
    Upload: "File",                        // File upload
    UUID: "string",                        // UUID string
    URL: "string",                         // URL string
    EmailAddress: "string",                // Email string
    PositiveInt: "number",                 // Positive integer
  },
}
```

## CI Integration

```yaml
# .github/workflows/graphql.yml
name: GraphQL Checks
on: [push, pull_request]

jobs:
  graphql:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: 20 }
      - run: npm ci

      # Check that generated types are up to date
      - run: npx graphql-codegen
      - run: git diff --exit-code src/gql/
        # Fails if codegen output differs from committed files

      # Validate operations against schema
      - run: npx graphql-inspector validate './src/**/*.{ts,tsx}' ./schema.graphql
```

## Gotchas

1. **Import `gql` from generated output, not `@apollo/client`** — The client preset generates a typed `gql` function at `./src/gql/`. Using the untyped `gql` from `@apollo/client` defeats the purpose of codegen.

2. **Run codegen before TypeScript compilation** — Generated types must exist before `tsc` runs. Add `"codegen"` as a `prebuild` script or to your CI pipeline.

3. **Fragment masking requires `getFragmentData`** — The client preset uses fragment masking by default. You can't directly access fragment fields without calling `getFragmentData()`. This enforces component data boundaries.

4. **Schema changes require re-running codegen** — After modifying your GraphQL schema, run `npx graphql-codegen` to regenerate types. Watch mode (`--watch`) handles this automatically during development.

5. **Don't edit generated files** — Files in `./src/gql/` are overwritten on every codegen run. Put custom types in separate files.

6. **Mapper types prevent resolver type mismatches** — Without mappers, resolver return types match GraphQL types exactly. With mappers, they match your DB model types, and field resolvers bridge the gap. Configure mappers for any type where the DB model differs from the GraphQL type.
