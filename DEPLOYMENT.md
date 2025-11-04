# Deployment Guide

This guide covers how to deploy DocuFlow AI to production. The application consists of two parts:
- **Backend**: Go API server (port 8080)
- **Frontend**: React/Vite application

## Quick Start Options

### Option 1: Render (Backend) + Vercel (Frontend) ⭐ Recommended
- **Backend**: Render.com (free tier available)
- **Frontend**: Vercel (free tier available)
- **Best for**: Quick deployment with minimal configuration

### Option 2: Railway (Both Services)
- **Both**: Railway.app (pay-as-you-go)
- **Best for**: Simpler setup with both services in one place

### Option 3: Docker + Any Platform
- **Backend**: Docker container
- **Frontend**: Static hosting (Vercel, Netlify, etc.)
- **Best for**: Maximum flexibility

---

## Option 1: Render + Vercel (Recommended)

### Step 1: Deploy Backend to Render

1. **Sign up** at [render.com](https://render.com)

2. **Create a new Web Service:**
   - Click "New +" → "Web Service"
   - Connect your GitHub repository
   - Select the repository

3. **Configure the service:**
   ```
   Name: docuflow-backend
   Environment: Go
   Build Command: cd server && go mod download && go build -o app
   Start Command: ./app
   ```

4. **Set Environment Variables:**
   ```
   GEMINI_API_KEY=your_gemini_api_key_here
   PORT=8080
   ENV=production
   ```

5. **Update CORS in `server/main.go`** (see instructions below)

6. **Deploy**: Click "Create Web Service"

7. **Note the URL**: You'll get a URL like `https://docuflow-backend.onrender.com`

### Step 2: Deploy Frontend to Vercel

1. **Sign up** at [vercel.com](https://vercel.com)

2. **Import your repository:**
   - Click "New Project"
   - Import your GitHub repository

3. **Configure the project:**
   ```
   Framework Preset: Vite
   Root Directory: client
   Build Command: npm run build
   Output Directory: dist
   ```

4. **Set Environment Variables:**
   ```
   VITE_API_URL=https://your-backend-url.onrender.com/api
   ```
   (Replace with your actual Render backend URL)

5. **Deploy**: Click "Deploy"

6. **Your app is live!** You'll get a URL like `https://docuflow-ai.vercel.app`

---

## Option 2: Railway (Both Services)

### Backend Deployment

1. **Sign up** at [railway.app](https://railway.app)

2. **Create a new project** → "Deploy from GitHub repo"

3. **Add a service** → Select your repo

4. **Configure:**
   - Railway will auto-detect Go
   - Set **Root Directory**: `server`
   - Set **Start Command**: `go run main.go` (or build: `go build -o app && ./app`)

5. **Add Environment Variables:**
   ```
   GEMINI_API_KEY=your_gemini_api_key_here
   ENV=production
   ```
   (Port is auto-assigned)

6. **Deploy**: Railway will auto-deploy

### Frontend Deployment

1. **Add another service** in the same Railway project

2. **Select your repo again**

3. **Configure:**
   - Set **Root Directory**: `client`
   - Set **Build Command**: `npm install && npm run build`
   - Set **Start Command**: `npx serve -s dist -p $PORT`

4. **Add Environment Variables:**
   ```
   VITE_API_URL=https://your-backend-service.railway.app/api
   ```

5. **Add Nixpacks config** (Railway will auto-detect, but you can add `client/nixpacks.toml`):
   ```toml
   [phases.setup]
   nixPkgs = ["nodejs-18_x"]

   [phases.build]
   cmds = ["npm install", "npm run build"]

   [start]
   cmd = "npx serve -s dist -p $PORT"
   ```

---

## Option 3: Docker Deployment

### Step 1: Create Dockerfile for Backend

Create `server/Dockerfile`:
```dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
```

### Step 2: Create Dockerfile for Frontend

Create `client/Dockerfile`:
```dockerfile
FROM node:18-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html

COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

Create `client/nginx.conf`:
```nginx
server {
    listen 80;
    server_name _;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

### Step 3: Deploy Docker Containers

**Using Docker Compose** (for local/dedicated server):
```yaml
# docker-compose.yml
version: '3.8'

services:
  backend:
    build: ./server
    ports:
      - "8080:8080"
    environment:
      - GEMINI_API_KEY=${GEMINI_API_KEY}
      - PORT=8080
      - ENV=production
    restart: unless-stopped

  frontend:
    build: 
      context: ./client
      args:
        - VITE_API_URL=http://localhost:8080/api
    ports:
      - "80:80"
    depends_on:
      - backend
    restart: unless-stopped
```

**Using Cloud Platforms:**
- **Google Cloud Run**: Push Docker images and deploy
- **AWS ECS/Fargate**: Use Dockerfiles above
- **Azure Container Instances**: Same Dockerfiles work

---

## Important: Update CORS Configuration

Before deploying, you **must** update CORS settings in `server/main.go` to allow your production frontend URL:

```go
// Update this section in server/main.go
r.Use(cors.New(cors.Config{
    AllowOrigins: []string{
        "http://localhost:5173",      // Keep for local dev
        "http://localhost:3000",      // Keep for local dev
        "https://your-frontend.vercel.app",  // Add your production URL
        "https://your-frontend.railway.app", // Or Railway URL
    },
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
    ExposeHeaders:    []string{"Content-Disposition"},
    AllowCredentials: true,
    MaxAge:           300,
}))
```

**Better approach** (use environment variable):
```go
allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
if allowedOrigins == "" {
    allowedOrigins = "http://localhost:5173,http://localhost:3000"
}
origins := strings.Split(allowedOrigins, ",")

r.Use(cors.New(cors.Config{
    AllowOrigins:     origins,
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
    ExposeHeaders:    []string{"Content-Disposition"},
    AllowCredentials: true,
    MaxAge:           300,
}))
```

Then set in production:
```
ALLOWED_ORIGINS=https://your-frontend.vercel.app,https://your-frontend.railway.app
```

---

## Environment Variables Checklist

### Backend (Production)
```
GEMINI_API_KEY=your_api_key_here
PORT=8080 (or auto-assigned by platform)
ENV=production
ALLOWED_ORIGINS=https://your-frontend-url.com
```

### Frontend (Production)
```
VITE_API_URL=https://your-backend-url.com/api
```

---

## Testing Your Deployment

1. **Backend Health Check:**
   ```bash
   curl https://your-backend-url.com/api/health
   ```
   Should return: `{"status":"ok"}`

2. **Frontend:**
   - Visit your frontend URL
   - Try uploading a document
   - Check browser console for errors
   - Verify API calls are going to correct backend URL

3. **CORS Issues:**
   - If you see CORS errors, double-check `ALLOWED_ORIGINS`
   - Ensure frontend URL matches exactly (including https/http)

---

## Platform-Specific Notes

### Render
- Free tier sleeps after 15 minutes of inactivity
- Paid tier: $7/month for always-on
- Good for backend APIs

### Vercel
- Free tier is excellent for frontend
- Automatic HTTPS
- Global CDN
- Great for React/Vite apps

### Railway
- Pay-as-you-go pricing
- Easy deployment
- Good for full-stack apps
- Auto-sleeps on free tier

### Heroku (Alternative)
- Requires credit card even for free tier
- Buildpack-based deployment
- Good alternative if others don't work

---

## Troubleshooting

### Backend won't start
- Check `PORT` environment variable
- Verify `GEMINI_API_KEY` is set
- Check logs in platform dashboard

### Frontend can't connect to backend
- Verify `VITE_API_URL` is correct
- Check CORS settings in backend
- Ensure backend is running and accessible

### CORS errors
- Double-check `ALLOWED_ORIGINS` includes your frontend URL
- Verify URL matches exactly (protocol, domain, port)
- Check browser console for specific error

### Build failures
- Check platform logs for specific errors
- Verify all dependencies are in `package.json` / `go.mod`
- Ensure build commands are correct

---

## Next Steps

After deployment:
1. ✅ Set up custom domain (optional)
2. ✅ Enable HTTPS (usually automatic)
3. ✅ Set up monitoring/logging
4. ✅ Configure backups (if using database later)
5. ✅ Set up CI/CD for automatic deployments

---

## Support

If you encounter issues:
1. Check platform logs
2. Verify environment variables
3. Test locally with production-like settings
4. Check this guide for common issues

