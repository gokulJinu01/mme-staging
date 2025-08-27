# Traefik ForwardAuth Setup for MME

## Overview

This setup uses Traefik API Gateway with ForwardAuth to handle authentication and cross-tenant isolation at the gateway level, simplifying the service code.

## Architecture

```
Client → Traefik → JWT Verifier → MME Services
```

### Components

1. **Traefik API Gateway**: Routes requests and applies middleware
2. **JWT Verifier**: Validates JWT tokens and sets user headers
3. **MME Services**: Focus on business logic, read user from headers

## Services

### Traefik (`traefik.localhost:8080`)
- **Dashboard**: http://traefik.localhost:8080
- **Entry Point**: http://localhost (port 80)

### JWT Verifier (`auth.localhost`)
- **Health Check**: http://auth.localhost/health
- **ForwardAuth**: http://auth.localhost/verify

### MME Tagging Service (`mme.localhost`)
- **API**: http://mme.localhost/memory/*
- **Health**: http://mme.localhost/health

### MME Tagmaker Service (`tagmaker.localhost`)
- **API**: http://tagmaker.localhost/*
- **Health**: http://tagmaker.localhost/health

## Authentication Flow

1. **Client Request**: `GET /memory/query?userId=user123`
2. **Traefik**: Forwards to JWT Verifier for auth check
3. **JWT Verifier**: Validates token/headers, sets `X-User-ID`
4. **MME Service**: Reads `X-User-ID` header, processes request

## Headers Set by JWT Verifier

- `X-User-ID`: User identifier
- `X-User-Roles`: User roles (JSON array)
- `X-Org-ID`: Organization identifier
- `X-Project-ID`: Project identifier

## Testing

### Development Mode
For development/testing, the JWT verifier accepts requests with `X-User-ID` header:

```bash
curl -H "X-User-ID: test_user" http://mme.localhost/memory/query
```

### Production Mode
For production, use JWT tokens:

```bash
curl -H "Authorization: Bearer <jwt_token>" http://mme.localhost/memory/query
```

## Benefits

1. **Centralized Auth**: All services use same auth logic
2. **Cross-Tenant Isolation**: Enforced at gateway level
3. **Simplified Service Code**: Services focus on business logic
4. **Better Performance**: Auth decisions made once
5. **Consistent Security**: Same rules across all endpoints

## Configuration Files

- `traefik.yml`: Traefik configuration
- `docker-compose.yml`: Service definitions
- `jwt-verifier/`: JWT verification service

## Troubleshooting

### Check Traefik Dashboard
Visit http://traefik.localhost:8080 to see:
- Service health
- Request routing
- Middleware configuration

### Check JWT Verifier Logs
```bash
docker-compose logs jwt-verifier
```

### Test Authentication
```bash
./test-traefik-setup.sh
```

## Migration from Service-Level Auth

1. ✅ Removed JWT middleware from services
2. ✅ Removed auth middleware from routes
3. ✅ Simplified handlers to read `X-User-ID` header
4. ✅ Added Traefik ForwardAuth configuration
5. ✅ Created JWT verifier service
6. ✅ Updated docker-compose.yml

## Next Steps

1. **Test the setup**: Run `./test-traefik-setup.sh`
2. **Verify cross-tenant isolation**: Test with different user IDs
3. **Monitor performance**: Check Traefik dashboard
4. **Deploy to production**: Update JWT secret and CORS settings
