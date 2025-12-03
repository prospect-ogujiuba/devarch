import express from 'express'
import cors from 'cors'
import helmet from 'helmet'
import morgan from 'morgan'
import path from 'path'
import { fileURLToPath } from 'url'
import dotenv from 'dotenv'
import apiRoutes from './routes/api.js'

// Load environment variables
dotenv.config()

// Get __dirname equivalent in ES modules
const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const app = express()
const PORT = process.env.PORT || 8200

// Middleware
app.use(helmet())
app.use(cors())
app.use(morgan('dev'))
app.use(express.json())
app.use(express.urlencoded({ extended: true }))

// Serve static files from public/ directory - CRITICAL for DevArch
app.use(express.static(path.join(__dirname, '../public')))

// API routes
app.use('/api', apiRoutes)

// Health check endpoint
app.get('/api/health', (req, res) => {
  res.json({
    status: 'ok',
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
  })
})

// SPA fallback - serve index.html for all other routes
app.get('*', (req, res) => {
  res.sendFile(path.join(__dirname, '../public/index.html'))
})

// Error handling middleware
app.use((err, req, res, next) => {
  console.error(err.stack)
  res.status(500).json({
    error: 'Internal Server Error',
    message: process.env.NODE_ENV === 'development' ? err.message : undefined,
  })
})

// Start server
app.listen(PORT, '0.0.0.0', () => {
  console.log(`
    🚀 DevArch Express Server Running
    ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    Port: ${PORT}
    Environment: ${process.env.NODE_ENV || 'development'}
    Static files: ${path.join(__dirname, '../public')}
    ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  `)
})

export default app
