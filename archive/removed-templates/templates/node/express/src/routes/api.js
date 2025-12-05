import express from 'express'

const router = express.Router()

// Example API routes
router.get('/', (req, res) => {
  res.json({
    message: 'DevArch Express API',
    version: '1.0.0',
    endpoints: {
      health: '/api/health',
      users: '/api/users',
    },
  })
})

router.get('/users', (req, res) => {
  res.json({
    users: [
      { id: 1, name: 'John Doe', email: 'john@example.com' },
      { id: 2, name: 'Jane Smith', email: 'jane@example.com' },
    ],
  })
})

router.post('/users', (req, res) => {
  const { name, email } = req.body
  res.status(201).json({
    id: Date.now(),
    name,
    email,
    createdAt: new Date().toISOString(),
  })
})

export default router
