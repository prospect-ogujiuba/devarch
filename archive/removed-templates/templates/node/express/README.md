# DevArch Express.js Template

An Express.js server template configured to serve static files from `public/` directory with built-in API support, following DevArch's standardized structure.

## Features

- Express.js 4+ server
- Static file serving from `public/` directory
- Built-in REST API routes
- Security middleware (Helmet, CORS)
- Request logging (Morgan)
- Environment variable configuration
- Hot-reload with Nodemon
- **Serves static files from `public/` for web server compatibility**

## Quick Start

### 1. Install Dependencies

```bash
npm install
```

### 2. Configure Environment

```bash
cp .env.example .env
# Edit .env with your configuration
```

### 3. Development Mode

```bash
npm run dev
```

Access at: http://localhost:8200

### 4. Production Mode

```bash
npm start
```

## Directory Structure

```
express-app/
├── public/              # WEB ROOT - Static files served from here
│   ├── index.html       # Main HTML file
│   ├── assets/          # CSS, JS, images
│   └── api/             # Optional: Static API documentation
├── src/
│   ├── server.js        # Main Express server
│   ├── routes/          # API routes
│   │   └── api.js
│   ├── controllers/     # Request controllers
│   ├── middleware/      # Custom middleware
│   └── config/          # Configuration files
├── package.json
└── .env.example
```

## Static Files Configuration

The server is configured to serve files from `public/`:

```javascript
// src/server.js
app.use(express.static(path.join(__dirname, '../public')))
```

**This is critical** for DevArch compatibility. All static assets must be in `public/`.

## API Routes

API endpoints are prefixed with `/api`:

- `GET /api` - API information
- `GET /api/health` - Health check
- `GET /api/users` - List users (example)
- `POST /api/users` - Create user (example)

### Adding New Routes

Create route file in `src/routes/`:

```javascript
// src/routes/myroute.js
import express from 'express'
const router = express.Router()

router.get('/', (req, res) => {
  res.json({ message: 'My route' })
})

export default router
```

Register in `src/server.js`:

```javascript
import myRoutes from './routes/myroute.js'
app.use('/api/my', myRoutes)
```

## Environment Variables

```env
NODE_ENV=development
PORT=8200
CORS_ORIGIN=*
```

Access in code:

```javascript
const port = process.env.PORT || 8200
```

## DevArch Integration

### Port Assignment

Node.js apps use ports 8200-8299:

```env
PORT=8200
```

### Domain Configuration

Configure Nginx Proxy Manager:

1. Access: http://localhost:81
2. Add proxy host for `your-app.test`
3. Point to: `http://nodejs:8200`
4. Enable SSL certificate

### Container Access

- Development: http://localhost:8200
- Production: https://your-app.test

## Middleware Stack

The server includes:

- **helmet**: Security headers
- **cors**: Cross-origin resource sharing
- **morgan**: Request logging
- **express.json()**: JSON body parsing
- **express.static()**: Static file serving

## Error Handling

Global error handler:

```javascript
app.use((err, req, res, next) => {
  console.error(err.stack)
  res.status(500).json({
    error: 'Internal Server Error',
    message: process.env.NODE_ENV === 'development' ? err.message : undefined,
  })
})
```

## SPA Support

The server includes SPA fallback routing:

```javascript
app.get('*', (req, res) => {
  res.sendFile(path.join(__dirname, '../public/index.html'))
})
```

This serves `index.html` for all non-API routes, enabling client-side routing.

## Testing

Add testing dependencies:

```bash
npm install --save-dev jest supertest
```

Create test files:

```javascript
// __tests__/api.test.js
import request from 'supertest'
import app from '../src/server.js'

test('GET /api/health returns ok', async () => {
  const response = await request(app).get('/api/health')
  expect(response.status).toBe(200)
  expect(response.body.status).toBe('ok')
})
```

Run tests:

```bash
npm test
```

## Database Integration

### Example with MySQL/MariaDB

```bash
npm install mysql2
```

```javascript
// src/config/database.js
import mysql from 'mysql2/promise'

const pool = mysql.createPool({
  host: process.env.DB_HOST,
  user: process.env.DB_USER,
  password: process.env.DB_PASSWORD,
  database: process.env.DB_NAME,
})

export default pool
```

## Troubleshooting

### Port Already in Use

Change port in `.env`:

```env
PORT=8201
```

### Static Files Not Loading

Verify `public/` directory exists and path is correct in server.js.

### CORS Errors

Update CORS origin in `.env`:

```env
CORS_ORIGIN=https://your-frontend.test
```

## Learn More

- [Express.js Documentation](https://expressjs.com/)
- [Node.js Best Practices](https://github.com/goldbergyoni/nodebestpractices)
- [DevArch Documentation](../../APP_STRUCTURE.md)

## Support

For template-specific issues:
- Check DevArch documentation: `APP_STRUCTURE.md`
- Review template documentation: `templates/README.md`
- Verify static files are in `public/` directory
