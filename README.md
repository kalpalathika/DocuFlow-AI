# DocuFlow AI

An AI-powered legal document automation platform that simplifies document filling through conversational interfaces. Upload a template document, answer questions naturally, and generate completed documents automatically.

## Overview

DocuFlow AI provides a seamless document automation experience by:
- Intelligently detecting placeholders in `.docx` templates using AI
- Generating natural, conversational questions from technical field names
- Guiding users through a question-answer workflow
- Automatically filling documents with user responses
- Supporting multiple placeholder formats: `{{field_name}}`, `[Field Name]`, `$[__________]`

## Project Structure

```
DocuFlow-AI/
├── server/                 # Go backend service
│   ├── handlers/          # HTTP request handlers
│   ├── docx/              # Document processing logic
│   ├── models/            # Data models and types
│   ├── session/           # Session management
│   ├── utils/             # Utility functions
│   └── main.go            # Application entry point
├── client/                 # React frontend application
│   ├── src/
│   │   ├── pages/         # Page components
│   │   ├── lib/           # API client and utilities
│   │   └── assets/        # Static assets
│   └── package.json
└── README.md
```

## Technologies Used

### Backend

Go (v1.23), Gin Framework, nguyenthenguyen/docx, Gemini API (Google)

### Frontend

React 19, TypeScript, Vite, Tailwind CSS, React Router

## Features

- **Smart Field Detection**: AI-powered identification of placeholders in documents using Gemini API
- **Conversational Interface**: Natural language questions generated from technical field names
- **Multiple Input Types**: Support for text, number, and date fields with appropriate UI controls
- **Progress Tracking**: Real-time progress indicators showing completion status
- **Document Generation**: Automatically fill templates with user responses
- **Error Handling**: Graceful handling of API quota limits with user-friendly messages
- **Session Management**: In-memory session storage for document filling workflows

## Getting Started

### Prerequisites

- **Go** (v1.23 or higher)
- **Node.js** (v18 or higher)
- **npm** or **yarn**
- **Gemini API Key** (for field detection)

### Installation

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd DocuFlow-AI
   ```

2. **Set up the backend:**
   ```bash
   cd server
   go mod download
   ```

3. **Set up the frontend:**
   ```bash
   cd ../client
   npm install
   ```

4. **Configure environment variables:**

   For the backend (`server/.env` or export variables):
   ```bash
   export GEMINI_API_KEY=your_gemini_api_key_here
   export PORT=8080
   export ENV=development
   ```

   For the frontend (`client/.env`):
   ```bash
   VITE_API_URL=http://localhost:8080/api
   ```

### Running the Application

**Backend:**
```bash
cd server
go run main.go
```

The API server will start on `http://localhost:8080`

**Frontend:**
```bash
cd client
npm run dev
```

The frontend will start on `http://localhost:5173`

## Development

### Backend Development

Run tests:
```bash
cd server
go test ./...
```

Run with verbose test output:
```bash
go test ./handlers -v
```

### Frontend Development

Run in development mode:
```bash
cd client
npm run dev
```

Build for production:
```bash
npm run build
```

Preview production build:
```bash
npm run preview
```

## API Endpoints

### Health Check
- **GET** `/api/health`
- Returns server status

### Document Upload
- **POST** `/api/upload`
- Upload a `.docx` template file
- Form data field: `document` or `file`
- Returns: `{ sessionId, fields[], message }`

### Session Management
- **GET** `/api/session/:id`
- Get session status and current answers
- Returns: `{ sessionId, fields[], answers{}, questions{}, progress, total, isCompleted }`

### Questions & Answers
- **GET** `/api/session/:id/next`
- Get the next unanswered question
- Returns: `{ field, fieldType, question, isAIPhrased, progress, total, done }`

- **POST** `/api/session/:id/answers`
- Submit an answer for a field
- Body: `{ field: string, answer: string }`
- Returns: `{ message, field, progress, total }`

### AI Enhancement
- **POST** `/api/session/:id/ai/questions`
- Generate AI-phrased questions for all fields (optional)
- Returns: `{ questions{}, count, message }`

### Document Generation
- **POST** `/api/session/:id/generate`
- Generate the filled document for download
- Returns: DOCX file download

## Design Decisions

### Separation of Concerns
- **Handlers**: HTTP request/response handling only
- **Business Logic**: Document processing and field detection in separate packages
- **Models**: Clean data structures shared across packages
- **Session Store**: Thread-safe in-memory storage with mutex protection

### AI-Powered Field Detection
- Uses Gemini API to intelligently detect placeholders from document context
- Distinguishes between dynamic placeholders and static references (e.g., `[Section 1]` vs `[Company Name]`)
- Falls back to simple pattern matching if AI detection fails

### Error Handling
- Comprehensive error detection for Gemini API quota exhaustion
- User-friendly error messages with actionable guidance
- Graceful degradation when external services are unavailable

### Session Management
- In-memory storage for simplicity (suitable for MVP)
- Thread-safe concurrent access using read-write mutexes
- Session-based workflow ensures data integrity

### Frontend Architecture
- Type-safe API client with matching backend types
- Component-based UI for reusability
- Progressive form filling with real-time feedback

## Testing

Integration tests are included in `server/handlers/handlers_test.go`:

- **TestHealthCheck**: Verifies health endpoint functionality
- **TestUploadInvalidFileType**: Tests file validation
- **TestSessionWorkflow**: Complete end-to-end session workflow test

Run tests:
```bash
cd server
go test ./handlers -v
```

## Challenges & Solutions

### Challenge: Gemini API Quota Limits
**Solution**: Implemented comprehensive error detection for quota exhaustion with user-friendly error messages prompting users to retry after a short delay.

### Challenge: Field Detection Accuracy
**Solution**: AI-powered detection using Gemini API with intelligent context analysis to distinguish between placeholders and static document references.

### Challenge: Document Format Support
**Solution**: Focused on `.docx` format with robust parsing using the nguyenthenguyen/docx library, providing wide compatibility with Microsoft Word documents.

## Future Enhancements

### Short-term
- **Database Integration**: Replace in-memory storage with PostgreSQL or MongoDB for persistent session storage
- **User Authentication**: Add JWT-based authentication for multi-user support
- **Enhanced Error Handling**: Structured logging and error tracking (e.g., Sentry)
- **API Rate Limiting**: Implement rate limiting to prevent abuse

### Medium-term
- **Document Templates Library**: Pre-built template collection , smart OCR Detection
- **Export Formats**: Support for PDF, HTML exports
- **Batch Processing**: Handle multiple documents simultaneously
- **Collaboration Features**: Share documents with team members

### Long-term
- **Cloud Storage Integration**: Direct integration with Google Drive, Dropbox
- **Version Control**: Track document revisions and changes
- **Advanced AI Features**: 
  - Document summarization
  - Smart field suggestions
  - Legal clause recommendations
- **Microservices Architecture**: Split into separate services for scalability

## Making it Production-Ready

To prepare for production deployment:

1. **Environment Management**: Use environment variables for all configuration
2. **Security**: 
   - Implement authentication and authorization
   - Add input validation and sanitization
   - Enable HTTPS/TLS
3. **Monitoring**: 
   - Set up logging (structured logging with levels)
   - Add health check endpoints
   - Implement metrics collection (Prometheus/Grafana)
4. **Database**: Migrate from in-memory to persistent database
5. **Caching**: Add Redis for frequently accessed data
6. **CI/CD**: Set up automated testing and deployment pipelines
7. **Documentation**: API documentation (Swagger/OpenAPI)

