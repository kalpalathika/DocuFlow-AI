// API client configuration and utilities

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

// Types matching backend models
export interface Session {
  id: string;
  fields: string[];
  answers: Record<string, string>;
  questions: Record<string, string>;
  createdAt: string;
  updatedAt: string;
}

export interface UploadResponse {
  sessionId: string;
  fields: string[];
  message: string;
}

export interface QuestionResponse {
  field: string;
  question: string;
  isAIPhrased: boolean;
  progress: number;
  total: number;
  done: boolean;
}

export interface SessionStatusResponse {
  sessionId: string;
  fields: string[];
  answers: Record<string, string>;
  questions: Record<string, string>;
  progress: number;
  total: number;
  isCompleted: boolean;
}

export interface ErrorResponse {
  error: string;
  message?: string;
}

// API functions
export const api = {
  // Upload document
  async uploadDocument(file: File): Promise<UploadResponse> {
    const formData = new FormData();
    formData.append('document', file);

    const response = await fetch(`${API_BASE_URL}/upload`, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      const error: ErrorResponse = await response.json();
      throw new Error(error.message || 'Upload failed');
    }

    return response.json();
  },

  // Get session status
  async getSession(sessionId: string): Promise<SessionStatusResponse> {
    const response = await fetch(`${API_BASE_URL}/session/${sessionId}`);

    if (!response.ok) {
      const error: ErrorResponse = await response.json();
      throw new Error(error.message || 'Failed to get session');
    }

    return response.json();
  },

  // Get next question
  async getNextQuestion(sessionId: string): Promise<QuestionResponse> {
    const response = await fetch(`${API_BASE_URL}/session/${sessionId}/next`);

    if (!response.ok) {
      const error: ErrorResponse = await response.json();
      throw new Error(error.message || 'Failed to get next question');
    }

    return response.json();
  },

  // Submit answer
  async submitAnswer(sessionId: string, field: string, answer: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/session/${sessionId}/answers`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ field, answer }),
    });

    if (!response.ok) {
      const error: ErrorResponse = await response.json();
      throw new Error(error.message || 'Failed to submit answer');
    }
  },

  // Generate AI questions (optional)
  async generateAIQuestions(sessionId: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/session/${sessionId}/ai/questions`, {
      method: 'POST',
    });

    if (!response.ok) {
      const error: ErrorResponse = await response.json();
      throw new Error(error.message || 'Failed to generate AI questions');
    }
  },

  // Download filled document
  async downloadDocument(sessionId: string): Promise<Blob> {
    const response = await fetch(`${API_BASE_URL}/session/${sessionId}/generate`, {
      method: 'POST',
    });

    if (!response.ok) {
      const error: ErrorResponse = await response.json();
      throw new Error(error.message || 'Failed to generate document');
    }

    return response.blob();
  },
};
