# WebStorm - Express/Node Backend Quick Start

Create Express.js backend applications in DevArch using WebStorm with container-based Node.js runtime.

## Prerequisites

- WebStorm installed
- DevArch containers running:
  ```bash
  ./scripts/service-manager.sh start database proxy backend
  ```

## 1. Create Project

**Create directory and initialize:**

```bash
podman exec -it node bash
cd /app
mkdir my-express-api
cd my-express-api
npm init -y
npm install express cors dotenv
npm install -D nodemon
exit
```

**Open in WebStorm:**
1. File → Open → `/home/fhcadmin/projects/devarch/apps/my-express-api`

## 2. Setup `public/` Structure

Express serves static files from `public/`:

```
apps/my-express-api/
├── public/              # Static assets served by Express
│   ├── index.html       # Optional: API documentation page
│   └── assets/
├── src/
│   ├── server.js        # Main entry point
│   ├── routes/
│   ├── controllers/
│   ├── models/
│   └── middleware/
├── .env
└── package.json
```

## 3. Create Express Server

**Create `src/server.js`:**

```javascript
import express from 'express'
import cors from 'cors'
import path from 'path'
import { fileURLToPath } from 'url'
import dotenv from 'dotenv'

dotenv.config()

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const app = express()
const PORT = process.env.PORT || 3000

// Middleware
app.use(cors())
app.use(express.json())
app.use(express.urlencoded({ extended: true }))

// Serve static files from public/
app.use(express.static(path.join(__dirname, '../public')))

// API routes
app.get('/api/health', (req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() })
})

app.get('/api/users', (req, res) => {
  res.json([
    { id: 1, name: 'Alice' },
    { id: 2, name: 'Bob' },
  ])
})

app.post('/api/users', (req, res) => {
  const { name, email } = req.body
  res.status(201).json({ id: 3, name, email })
})

// Catch-all for SPA (if serving frontend from public/)
app.get('*', (req, res) => {
  res.sendFile(path.join(__dirname, '../public/index.html'))
})

app.listen(PORT, '0.0.0.0', () => {
  console.log(`Server running on http://0.0.0.0:${PORT}`)
})
```

**Update `package.json`:**

```json
{
  "name": "my-express-api",
  "version": "1.0.0",
  "type": "module",
  "main": "src/server.js",
  "scripts": {
    "start": "node src/server.js",
    "dev": "nodemon src/server.js",
    "dev:debug": "nodemon --inspect=0.0.0.0:9229 src/server.js"
  },
  "dependencies": {
    "express": "^4.18.2",
    "cors": "^2.8.5",
    "dotenv": "^16.3.1"
  },
  "devDependencies": {
    "nodemon": "^3.0.1"
  }
}
```

## 4. Configure Node.js Interpreter

1. Settings → Languages & Frameworks → Node.js
2. Node interpreter: "..." → "+" → "Add Remote..."
3. Docker Compose:
   - Configuration file: `/home/fhcadmin/projects/devarch/compose/backend/node.yml`
   - Service: `node`
   - Node.js path: `/usr/local/bin/node`
4. Click "OK"

## 5. Configure Run Configurations

**Dev server with hot reload:**

1. Run → Edit Configurations → "+" → npm
2. Configuration:
   - Name: `dev (container)`
   - Package.json: `/home/fhcadmin/projects/devarch/apps/my-express-api/package.json`
   - Command: `run`
   - Scripts: `dev`
   - Node interpreter: Container interpreter
3. Click "OK"

**Debug configuration:**

1. Run → Edit Configurations → "+" → npm
2. Configuration:
   - Name: `dev:debug (container)`
   - Scripts: `dev:debug`
   - Node interpreter: Container interpreter
3. Click "OK"

## 6. Environment Variables

Create `.env`:
```env
PORT=3000
NODE_ENV=development
DATABASE_URL=postgresql://postgres:admin1234567@postgres:5432/my_express_db
REDIS_URL=redis://:admin1234567@redis:6379
JWT_SECRET=your-secret-key-here
```

**Load in server.js:**
```javascript
import dotenv from 'dotenv'
dotenv.config()

const dbUrl = process.env.DATABASE_URL
```

## 7. Development Workflow

**Start server:**

```bash
podman exec -it node bash
cd /app/my-express-api
npm run dev
```

Or use WebStorm run configuration.

**Server runs on:** http://localhost:8200

**Test endpoints:**
```bash
curl http://localhost:8200/api/health
curl http://localhost:8200/api/users
```

## 8. Database Integration (PostgreSQL)

**Install pg (PostgreSQL client):**

```bash
npm install pg
```

**Create database connection:** `src/db/pool.js`

```javascript
import pg from 'pg'
const { Pool } = pg

const pool = new Pool({
  host: 'postgres',
  port: 5432,
  user: 'postgres',
  password: 'admin1234567',
  database: 'my_express_db',
})

pool.on('error', (err) => {
  console.error('Unexpected error on idle client', err)
  process.exit(-1)
})

export default pool
```

**Create database (run once):**

```bash
podman exec -it postgres bash
psql -U postgres -c "CREATE DATABASE my_express_db;"
exit
```

**Create table migration:** `src/db/migrations/001_users.sql`

```sql
CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Run migration:**

```javascript
// src/db/migrate.js
import pool from './pool.js'
import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

async function migrate() {
  const sql = fs.readFileSync(path.join(__dirname, 'migrations/001_users.sql'), 'utf8')
  await pool.query(sql)
  console.log('Migration complete')
  process.exit(0)
}

migrate()
```

Run: `node src/db/migrate.js`

**Use in routes:**

```javascript
// src/routes/users.js
import express from 'express'
import pool from '../db/pool.js'

const router = express.Router()

router.get('/', async (req, res) => {
  try {
    const result = await pool.query('SELECT * FROM users ORDER BY created_at DESC')
    res.json(result.rows)
  } catch (err) {
    console.error(err)
    res.status(500).json({ error: 'Database error' })
  }
})

router.post('/', async (req, res) => {
  const { name, email } = req.body
  try {
    const result = await pool.query(
      'INSERT INTO users (name, email) VALUES ($1, $2) RETURNING *',
      [name, email]
    )
    res.status(201).json(result.rows[0])
  } catch (err) {
    console.error(err)
    res.status(500).json({ error: 'Database error' })
  }
})

export default router
```

**Mount in server.js:**

```javascript
import userRoutes from './routes/users.js'
app.use('/api/users', userRoutes)
```

## 9. Redis Integration (Caching)

**Install Redis client:**

```bash
npm install redis
```

**Setup Redis client:** `src/cache/redis.js`

```javascript
import { createClient } from 'redis'

const client = createClient({
  url: 'redis://:admin1234567@redis:6379'
})

client.on('error', (err) => console.error('Redis Client Error', err))

await client.connect()

export default client
```

**Use for caching:**

```javascript
import redis from '../cache/redis.js'

router.get('/cached-data', async (req, res) => {
  const cacheKey = 'my-data'

  // Check cache first
  const cached = await redis.get(cacheKey)
  if (cached) {
    return res.json({ data: JSON.parse(cached), cached: true })
  }

  // Fetch from database
  const result = await pool.query('SELECT * FROM users')

  // Store in cache (expire in 60 seconds)
  await redis.setEx(cacheKey, 60, JSON.stringify(result.rows))

  res.json({ data: result.rows, cached: false })
})
```

## 10. Debugging

**Start server in debug mode:**

```bash
npm run dev:debug
```

**Attach WebStorm debugger:**

1. Run → Edit Configurations → "+" → "Attach to Node.js/Chrome"
2. Configuration:
   - Name: `Attach to Express`
   - Host: `localhost`
   - Port: `9229`
3. Click "OK"

**Or use WebStorm's automatic attachment:**

1. Run → Debug 'dev:debug (container)'
2. Set breakpoints in `server.js` or route files
3. Send request to API endpoint
4. Debugger pauses at breakpoint

**Inspect variables, step through code, evaluate expressions.**

## 11. Testing (Jest)

**Install Jest:**

```bash
npm install -D jest supertest @types/jest
```

**Configure Jest:** `jest.config.js`

```javascript
export default {
  testEnvironment: 'node',
  transform: {},
  moduleFileExtensions: ['js'],
  testMatch: ['**/__tests__/**/*.js', '**/?(*.)+(spec|test).js'],
}
```

**Update `package.json`:**

```json
"scripts": {
  "test": "NODE_OPTIONS=--experimental-vm-modules jest",
  "test:watch": "NODE_OPTIONS=--experimental-vm-modules jest --watch"
}
```

**Write tests:** `src/__tests__/server.test.js`

```javascript
import request from 'supertest'
import express from 'express'
// Import your app setup (without .listen())

describe('API Endpoints', () => {
  test('GET /api/health returns 200', async () => {
    const response = await request(app).get('/api/health')
    expect(response.statusCode).toBe(200)
    expect(response.body.status).toBe('ok')
  })

  test('GET /api/users returns array', async () => {
    const response = await request(app).get('/api/users')
    expect(response.statusCode).toBe(200)
    expect(Array.isArray(response.body)).toBe(true)
  })
})
```

Run: `npm test`

## 12. Configure nginx-proxy-manager

1. Open http://localhost:81
2. Proxy Hosts → Add Proxy Host
   - Domain: `my-express-api.test`
   - Forward to: `node:3000`
   - WebSockets: ✓
3. SSL: Request new certificate, Force SSL
4. Click "Save"

**Update /etc/hosts:**
```bash
sudo sh -c 'echo "127.0.0.1 my-express-api.test" >> /etc/hosts'
```

**Access:** https://my-express-api.test/api/health

## 13. API Documentation (Optional)

**Serve Swagger UI from `public/`:**

```bash
npm install swagger-ui-express swagger-jsdoc
```

**Setup Swagger:** `src/swagger.js`

```javascript
import swaggerJsdoc from 'swagger-jsdoc'
import swaggerUi from 'swagger-ui-express'

const options = {
  definition: {
    openapi: '3.0.0',
    info: {
      title: 'Express API',
      version: '1.0.0',
    },
  },
  apis: ['./src/routes/*.js'],
}

const specs = swaggerJsdoc(options)

export { specs, swaggerUi }
```

**Mount in server.js:**

```javascript
import { specs, swaggerUi } from './swagger.js'
app.use('/api-docs', swaggerUi.serve, swaggerUi.setup(specs))
```

**Access:** http://localhost:8200/api-docs

## Port Allocation

- **8200**: Express API server
- **8201**: Additional services
- **9229**: Node debugger

## Troubleshooting

**Issue:** Cannot connect to database
- Use container name (`postgres`, not `localhost`)
- Verify database container running: `./scripts/service-manager.sh status database`
- Check credentials in `.env`

**Issue:** Debugger not attaching
- Verify server started with `--inspect=0.0.0.0:9229`
- Check port 9229 exposed in compose file
- Try restarting debugger connection

**Issue:** Changes not reflected (nodemon not restarting)
- Check nodemon watching correct files
- Verify volume mount: `/apps` → `/app` in node.yml
- Restart container if volume mount broken

## Next Steps

- Add JWT authentication middleware
- Implement rate limiting (express-rate-limit)
- Setup request logging (morgan)
- Add input validation (express-validator)
- Configure HTTPS in production
- Setup PM2 for process management
- Add Helmet.js for security headers
