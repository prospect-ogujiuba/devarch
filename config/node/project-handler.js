const express = require('express');
const path = require('path');
const fs = require('fs');
const { spawn } = require('child_process');

const app = express();
const PORT = 3000;

// Store running projects
const runningProjects = new Map();

// Middleware to handle project routing
app.use((req, res, next) => {
    const projectName = req.headers['x-project-name'];
    const projectPath = req.headers['x-project-path'];
    
    if (projectName && projectPath && fs.existsSync(projectPath)) {
        const packageJsonPath = path.join(projectPath, 'package.json');
        
        if (fs.existsSync(packageJsonPath)) {
            try {
                const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
                
                // Check if it's a server project
                if (packageJson.scripts && (packageJson.scripts.start || packageJson.scripts.dev)) {
                    // Try to start project if not already running
                    if (!runningProjects.has(projectName)) {
                        startProject(projectName, projectPath, packageJson);
                    }
                    
                    // Proxy to the project
                    const projectPort = 3000 + Math.abs(hashCode(projectName) % 1000);
                    proxyToProject(req, res, projectPort);
                    return;
                }
            } catch (error) {
                console.error('Error processing project:', error);
            }
        }
    }
    
    next();
});

function hashCode(str) {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
        const char = str.charCodeAt(i);
        hash = ((hash << 5) - hash) + char;
        hash = hash & hash; // Convert to 32-bit integer
    }
    return hash;
}

function startProject(projectName, projectPath, packageJson) {
    const projectPort = 3000 + Math.abs(hashCode(projectName) % 1000);
    
    // Determine start command
    let command, args;
    if (packageJson.scripts.dev) {
        command = 'npm';
        args = ['run', 'dev'];
    } else if (packageJson.scripts.start) {
        command = 'npm';
        args = ['start'];
    } else {
        return;
    }
    
    // Set environment variables
    const env = {
        ...process.env,
        PORT: projectPort.toString(),
        NODE_ENV: 'development'
    };
    
    // Start the project
    const child = spawn(command, args, {
        cwd: projectPath,
        env: env,
        stdio: 'pipe'
    });
    
    child.stdout.on('data', (data) => {
        console.log(`[${projectName}] ${data}`);
    });
    
    child.stderr.on('data', (data) => {
        console.error(`[${projectName}] ${data}`);
    });
    
    child.on('close', (code) => {
        console.log(`[${projectName}] Process exited with code ${code}`);
        runningProjects.delete(projectName);
    });
    
    runningProjects.set(projectName, { child, port: projectPort });
    console.log(`Started project ${projectName} on port ${projectPort}`);
}

function proxyToProject(req, res, port) {
    // Simple proxy implementation
    const http = require('http');
    const options = {
        hostname: 'localhost',
        port: port,
        path: req.url,
        method: req.method,
        headers: req.headers
    };
    
    const proxyReq = http.request(options, (proxyRes) => {
        res.writeHead(proxyRes.statusCode, proxyRes.headers);
        proxyRes.pipe(res);
    });
    
    proxyReq.on('error', (err) => {
        res.writeHead(502);
        res.end(`<h1>Project Starting...</h1><p>Please wait while the project starts up, then refresh.</p><p>Error: ${err.message}</p>`);
    });
    
    req.pipe(proxyReq);
}

// Fallback handler
app.use('*', (req, res) => {
    res.send(`
        <h1>Node.js Project Handler</h1>
        <p>Project: ${req.headers['x-project-name'] || 'Unknown'}</p>
        <p>Path: ${req.headers['x-project-path'] || 'Unknown'}</p>
        <p>Running Projects: ${Array.from(runningProjects.keys()).join(', ') || 'None'}</p>
        <p>This project either hasn't started yet or doesn't have a proper entry point.</p>
        <p>Make sure your package.json has a "start" or "dev" script.</p>
    `);
});

app.listen(PORT, '0.0.0.0', () => {
    console.log(`Node.js project handler listening on port ${PORT}`);
});
