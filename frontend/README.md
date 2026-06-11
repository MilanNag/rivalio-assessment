# TaskFlow Frontend

Next.js (App Router) frontend for the TaskFlow task management application.
See the [root README](../README.md) for full setup instructions.

```bash
npm install
npm run dev            # http://localhost:3000 (expects API on :8090)
npm test               # Jest + React Testing Library
npm run test:coverage
npm run build
```

Set `NEXT_PUBLIC_API_URL` to point at a non-default backend URL.
