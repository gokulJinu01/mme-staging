# MME Frontend

A modern, responsive React frontend for the Memory Management Engine (MME) system.

## 🚀 Features

### Core User Pages
- **Dashboard** - System overview with quick stats and recent activity
- **Memory Management** - Save, view, edit, and delete memories with AI-powered tagging
- **Memory Promotion** - AI-powered memory retrieval with relevance scoring
- **Tag Management** - Manage and visualize the tag system
- **Search & Query** - Advanced search capabilities across the memory graph
- **Analytics** - System performance and usage analytics

### Admin Pages
- **Admin Dashboard** - System administration overview
- **System Management** - Advanced system operations and configuration
- **Monitoring** - Real-time metrics and debugging tools

### Authentication & Settings
- **Login** - Multi-tenant authentication with role-based access
- **Settings** - User preferences and system configuration

## 🛠 Tech Stack

- **React 18** - Modern React with hooks and functional components
- **React Router v6** - Client-side routing
- **React Query** - Server state management and caching
- **Tailwind CSS** - Utility-first CSS framework
- **Lucide React** - Beautiful, customizable icons
- **React Hook Form** - Performant forms with validation
- **React Hot Toast** - Elegant notifications
- **Framer Motion** - Smooth animations
- **Recharts** - Composable charting library

## 📦 Installation

1. **Install dependencies:**
   ```bash
   npm install
   ```

2. **Start development server:**
   ```bash
   npm start
   ```

3. **Build for production:**
   ```bash
   npm run build
   ```

## 🔧 Configuration

### Environment Variables

Create a `.env` file in the frontend directory:

```env
# API Configuration
REACT_APP_API_URL=http://localhost:80

# Feature Flags
REACT_APP_ENABLE_DEBUG=true
REACT_APP_ENABLE_ANALYTICS=true

# Authentication
REACT_APP_AUTH_PROVIDER=local
```

### API Integration

The frontend communicates with the MME backend through the `apiService`. Key endpoints:

- **Health Check**: `GET /health`
- **Memory Operations**: `POST /memory/save`, `GET /memory/query`, `POST /memory/promote`
- **Tag Operations**: `POST /tags/extract`, `POST /tags/query`
- **Search**: `POST /search/semantic`
- **Admin**: `GET /admin/stats`, `POST /admin/cleanup`

## 🎨 Design System

### Colors
- **Primary**: Blue gradient (`#3b82f6` to `#1e40af`)
- **Secondary**: Gray scale (`#64748b` to `#0f172a`)
- **Success**: Green (`#22c55e`)
- **Warning**: Orange (`#f59e0b`)
- **Error**: Red (`#ef4444`)

### Components
- **Buttons**: Primary, secondary, outline, ghost variants
- **Cards**: Consistent card layouts with headers, content, and footers
- **Forms**: Input fields, textareas, selects with validation
- **Badges**: Status indicators and tags
- **Tables**: Data tables with sorting and pagination

### Responsive Design
- Mobile-first approach
- Breakpoints: `sm` (640px), `md` (768px), `lg` (1024px), `xl` (1280px)
- Collapsible sidebar on mobile
- Touch-friendly interactions

## 📱 Pages Overview

### 1. Dashboard (`/dashboard`)
- System statistics cards
- Quick action buttons
- Recent memories preview
- System health status

### 2. Memory Management (`/memory`)
- Save new memories with AI tagging
- Search and filter existing memories
- Edit and delete operations
- Memory detail modals

### 3. Memory Promotion (`/promote`)
- Goal-based memory retrieval
- Tag-based filtering
- Relevance scoring display
- Spike-trace logging

### 4. Tag Management (`/tags`)
- Tag extraction from content
- Tag hierarchy visualization
- Tag relationships (edges)
- Usage statistics

### 5. Search & Query (`/search`)
- Semantic search interface
- Query by prompt
- Advanced filtering
- Search history

### 6. Analytics (`/analytics`)
- Memory growth trends
- Tag usage patterns
- Search performance
- User activity

### 7. Admin Dashboard (`/admin`)
- System statistics
- Health monitoring
- User management
- Configuration

### 8. System Management (`/admin/system`)
- Memory cleanup operations
- Edge pruning controls
- Feature flag management
- Backfill operations

### 9. Monitoring (`/admin/monitoring`)
- Real-time metrics
- Log viewing
- Performance monitoring
- Debug endpoints

### 10. Login (`/login`)
- Multi-tenant authentication
- Role-based access control
- Demo mode support

### 11. Settings (`/settings`)
- Profile management
- API key management
- Notification settings
- Organization settings

## 🔐 Authentication

The frontend uses a custom authentication system with:

- **JWT Tokens** - Stored in localStorage
- **Role-based Access** - User, Admin, System roles
- **Multi-tenant Support** - Organization-scoped data
- **ForwardAuth** - Handled by Traefik gateway

### User Roles
- **User**: Basic memory operations
- **Admin**: System management and monitoring
- **System**: Full system access and debugging

## 📊 State Management

### React Query
- Server state caching
- Background refetching
- Optimistic updates
- Error handling

### Local State
- Form state with React Hook Form
- UI state (modals, loading, etc.)
- User preferences

## 🎯 Performance

### Optimizations
- **Code Splitting** - Route-based lazy loading
- **Memoization** - React.memo and useMemo
- **Virtual Scrolling** - For large lists
- **Image Optimization** - WebP format support
- **Bundle Analysis** - Webpack bundle analyzer

### Monitoring
- **Error Boundaries** - Graceful error handling
- **Performance Metrics** - Core Web Vitals
- **Analytics** - User interaction tracking

## 🧪 Testing

### Test Setup
```bash
# Run tests
npm test

# Run tests with coverage
npm test -- --coverage

# Run tests in watch mode
npm test -- --watch
```

### Test Structure
- **Unit Tests** - Component and utility functions
- **Integration Tests** - API interactions
- **E2E Tests** - User workflows (Cypress)

## 🚀 Deployment

### Production Build
```bash
npm run build
```

### Docker Deployment
```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build
EXPOSE 3000
CMD ["npm", "start"]
```

### Environment Configuration
- **Development**: `.env.development`
- **Production**: `.env.production`
- **Staging**: `.env.staging`

## 🔧 Development

### Code Style
- **ESLint** - JavaScript linting
- **Prettier** - Code formatting
- **Husky** - Git hooks
- **Commitlint** - Conventional commits

### Development Tools
- **React DevTools** - Component inspection
- **Redux DevTools** - State debugging
- **Network Tab** - API monitoring

### Common Commands
```bash
# Start development server
npm start

# Run tests
npm test

# Build for production
npm run build

# Analyze bundle
npm run analyze

# Format code
npm run format

# Lint code
npm run lint
```

## 📚 API Documentation

### Authentication
All API requests include authentication headers:
```javascript
headers: {
  'Authorization': `Bearer ${token}`,
  'X-User-ID': userId
}
```

### Error Handling
- **401 Unauthorized** - Redirect to login
- **403 Forbidden** - Show access denied
- **404 Not Found** - Show not found page
- **500 Server Error** - Show error page

### Response Format
```javascript
{
  success: true,
  data: {...},
  message: "Success message"
}
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

### Development Guidelines
- Follow the existing code style
- Add tests for new features
- Update documentation
- Use conventional commits

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For support and questions:
- **Documentation**: Check the docs folder
- **Issues**: Create a GitHub issue
- **Discussions**: Use GitHub discussions

## 🔄 Updates

### Version History
- **v1.0.0** - Initial release with all 11 pages
- **v1.1.0** - Added analytics and monitoring
- **v1.2.0** - Enhanced search and filtering

### Migration Guide
See the migration guide for upgrading between versions.
