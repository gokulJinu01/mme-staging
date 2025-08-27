const express = require('express');
const jwt = require('jsonwebtoken');
const cors = require('cors');

const app = express();
const PORT = process.env.PORT || 3000;
const JWT_SECRET = process.env.JWT_SECRET || 'your-secret-key';

app.use(cors());
app.use(express.json());

// ForwardAuth endpoint for Traefik
app.get('/verify', (req, res) => {
    const authHeader = req.headers.authorization;
    const userHeader = req.headers['x-user-id'];
    
    // For development/testing: allow requests with X-User-ID header
    if (userHeader) {
        console.log(`ForwardAuth: Allowing request with X-User-ID: ${userHeader}`);
        res.setHeader('X-User-ID', userHeader);
        res.setHeader('X-User-Roles', '["USER"]');
        res.setHeader('X-Org-ID', userHeader);
        return res.status(200).send('OK');
    }
    
    // Check for JWT token
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
        console.log('ForwardAuth: No valid authorization header');
        return res.status(401).send('Unauthorized');
    }
    
    const token = authHeader.substring(7);
    
    try {
        const decoded = jwt.verify(token, JWT_SECRET);
        console.log(`ForwardAuth: Valid JWT for user: ${decoded.user_id || decoded.sub}`);
        
        // Set headers for the downstream service
        res.setHeader('X-User-ID', decoded.user_id || decoded.sub);
        res.setHeader('X-User-Roles', JSON.stringify(decoded.roles || ['USER']));
        res.setHeader('X-Org-ID', decoded.org_id || decoded.user_id);
        res.setHeader('X-Project-ID', decoded.project_id || '');
        
        res.status(200).send('OK');
    } catch (error) {
        console.log(`ForwardAuth: JWT verification failed: ${error.message}`);
        res.status(401).send('Unauthorized');
    }
});

// Health check endpoint
app.get('/health', (req, res) => {
    res.status(200).json({ status: 'healthy', service: 'jwt-verifier' });
});

app.listen(PORT, () => {
    console.log(`JWT Verifier service running on port ${PORT}`);
});
